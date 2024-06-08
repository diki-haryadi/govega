package fotel

import (
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const LocalsCtxKey = "otel-ctx"

type Filter func(c *fiber.Ctx) bool

// Config defines the config for middleware.
type Config struct {
	Tracer                trace.Tracer
	TracerStartAttributes []trace.SpanStartOption
	// SpanName is a template for span naming.
	// The scope is fiber context.
	SpanName     string
	LocalKeyName string
	Filters      []Filter //Blacklist filter
	Propagator   propagation.TextMapPropagator
}

// ConfigDefault is the default config
var ConfigDefault = Config{
	SpanName:     "{{ .Method }} {{ .Path }}",
	LocalKeyName: LocalsCtxKey,
	TracerStartAttributes: []trace.SpanStartOption{
		trace.WithSpanKind(trace.SpanKindServer),
	},
	Propagator: propagation.NewCompositeTextMapPropagator(b3.New(b3.WithInjectEncoding(b3.B3MultipleHeader)), propagation.TraceContext{}, propagation.Baggage{}),
}

// Helper function to set default values
func configDefault(config ...Config) Config {
	// Return default config if nothing provided
	if len(config) < 1 {
		return ConfigDefault
	}

	// Override default config
	cfg := config[0]

	if cfg.SpanName == "" {
		cfg.SpanName = ConfigDefault.SpanName
	}

	if cfg.LocalKeyName == "" {
		cfg.LocalKeyName = ConfigDefault.LocalKeyName
	}

	if len(cfg.TracerStartAttributes) == 0 {
		cfg.TracerStartAttributes = ConfigDefault.TracerStartAttributes
	}

	if cfg.Propagator == nil {
		cfg.Propagator = ConfigDefault.Propagator
	}

	return cfg
}

func concatSpanOptions(sources ...[]trace.SpanStartOption) []trace.SpanStartOption {
	var spanOptions []trace.SpanStartOption
	for _, source := range sources {
		spanOptions = append(spanOptions, source...)
	}

	return spanOptions
}

func FilterPathExact(paths ...string) Filter {
	return func(c *fiber.Ctx) bool {
		for _, p := range paths {
			if c.Path() == p {
				return true
			}
		}
		return false
	}
}
