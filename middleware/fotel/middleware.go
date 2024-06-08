package fotel

import (
	"bytes"
	"fmt"
	"net/http"
	"text/template"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
)

var Tracer = otel.Tracer("fiber-otel-router")

// New creates a new middleware handler
func New(config ...Config) fiber.Handler {
	// Set default config
	cfg := configDefault(config...)

	spanTmpl := template.Must(template.New("span").Parse(cfg.SpanName))

	// Return new handler
	return func(c *fiber.Ctx) error {

		for _, f := range cfg.Filters {
			if f(c) {
				// Simply pass through to the handler if a filter match
				return c.Next()
			}
		}

		// concat all span options, dynamic and static
		spanOptions := concatSpanOptions(
			[]trace.SpanStartOption{
				trace.WithAttributes(semconv.HTTPMethodKey.String(c.Method())),
				trace.WithAttributes(semconv.HTTPTargetKey.String(string(c.Request().RequestURI()))),
				trace.WithAttributes(semconv.HTTPRouteKey.String(c.Route().Path)),
				trace.WithAttributes(semconv.HTTPURLKey.String(c.OriginalURL())),
				trace.WithAttributes(semconv.NetHostIPKey.String(c.IP())),
				trace.WithAttributes(semconv.HTTPUserAgentKey.String(string(c.Request().Header.UserAgent()))),
				trace.WithAttributes(semconv.HTTPRequestContentLengthKey.Int(c.Request().Header.ContentLength())),
				trace.WithAttributes(semconv.HTTPSchemeKey.String(c.Protocol())),
				trace.WithAttributes(semconv.NetTransportTCP),
				trace.WithSpanKind(trace.SpanKindServer),
				// TODO:
				// - x-forwarded-for
				// -
			},
			cfg.TracerStartAttributes,
		)

		spanName := new(bytes.Buffer)
		err := spanTmpl.Execute(spanName, c)
		if err != nil {
			return fmt.Errorf("cannot execute span name template: %w", err)
		}

		h := make(http.Header)
		c.Request().Header.VisitAll(func(k, v []byte) {
			h.Add(string(k), string(v))
		})

		ctx := cfg.Propagator.Extract(c.Context(), propagation.HeaderCarrier(h))

		otelCtx, span := Tracer.Start(
			ctx,
			spanName.String(),
			spanOptions...,
		)

		c.Locals(LocalsCtxKey, otelCtx)
		defer span.End()

		err = c.Next()

		statusCode := c.Response().StatusCode()
		attrs := semconv.HTTPAttributesFromHTTPStatusCode(statusCode)
		spanStatus, spanMessage := semconv.SpanStatusFromHTTPStatusCode(statusCode)
		span.SetAttributes(attrs...)
		span.SetStatus(spanStatus, spanMessage)

		return err
	}
}
