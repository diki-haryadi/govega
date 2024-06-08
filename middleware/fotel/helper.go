package fotel

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/trace"
)

func FromCtx(ctx *fiber.Ctx) context.Context {
	otelCtx, ok := ctx.Locals(LocalsCtxKey).(context.Context)
	if !ok {
		return ctx.Context()
	}
	return otelCtx
}

func SpanFromCtx(ctx *fiber.Ctx) trace.Span {
	otelCtx := FromCtx(ctx)
	if otelCtx == nil {
		return trace.SpanFromContext(ctx.Context())
	}

	return trace.SpanFromContext(otelCtx)
}
