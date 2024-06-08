package tracing

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	golibconf "github.com/diki-haryadi/govega/config"
	"github.com/diki-haryadi/govega/env"
	"github.com/diki-haryadi/govega/log"
	"github.com/diki-haryadi/govega/util"
	"github.com/uber/jaeger-client-go/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	oteljaeger "go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	stdout "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
)

type tracerConfig struct {
	JaegerAgentHost    string  `json:"jaeger_agent_host"`
	JaegerAgentPort    string  `json:"jaeger_agent_port"`
	JaegerCollectorURL string  `json:"jaeger_url"`
	JaegerMode         string  `json:"jaeger_mode"`
	NewRelicKey        string  `json:"newrelic_apikey"`
	NewRelicURL        string  `json:"newrelic_url"`
	OtelMode           string  `json:"otel_agent_mode"`
	OtelEndpoint       string  `json:"otel_agent_endpoint"`
	SampleRate         float64 `json:"tracer_sample_rate"`
}

// InitFromEnv returns an instance of Jaeger Tracer that read from env
// Env example in GKE deployment
//   - name: JAEGER_AGENT_HOST
//     valueFrom:
//     fieldRef:
//     fieldPath: status.hostIP
//   - name: JAEGER_SERVICE_NAME
//     value: service_name
//   - name: JAEGER_SAMPLER_PARAM
//     value: "1"
//   - name: JAEGER_SAMPLER_TYPE
//     value: const
func InitFromEnv(service string) (opentracing.Tracer, io.Closer) {
	cfg, err := config.FromEnv()
	if err != nil {
		panic(fmt.Sprintf("ERROR: cannot init Jaeger: %v\n", err))
	}

	tracer, closer, err := cfg.NewTracer(config.Logger(jaeger.StdLogger))
	if err != nil {
		panic(fmt.Sprintf("ERROR: cannot init Jaeger: %v\n", err))
	}

	return tracer, closer
}

// InitLocal returns an instance of Jaeger Tracer that samples 100% of traces and logs all spans to stdout.
func InitLocal(service string) (opentracing.Tracer, io.Closer) {
	cfg := &config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans: true,
		},
	}
	tracer, closer, err := cfg.NewTracer(config.Logger(jaeger.StdLogger))
	if err != nil {
		panic(fmt.Sprintf("ERROR: cannot init Jaeger: %v\n", err))
	}
	return tracer, closer
}

func GetOtelProvider(kind, name string, conf interface{}) (trace.TracerProvider, error) {

	var cfg *tracerConfig

	if conf == nil {
		if err := golibconf.EnvToStruct(&cfg); err != nil {
			return nil, err
		}
	}

	switch cf := conf.(type) {
	case tracerConfig:
		cfg = &cf
	case *tracerConfig:
		cfg = cf
	case golibconf.Getter:
		cfg = new(tracerConfig)
		if err := cf.Unmarshal(cfg); err != nil {
			return nil, err
		}
	default:
		cfg = new(tracerConfig)
		if err := util.DecodeJSON(conf, cfg); err != nil {
			return nil, err
		}
	}

	if cfg == nil && kind != "no-op" {
		return nil, errors.New("missing configuration")
	}

	var provider trace.TracerProvider
	var exporter sdk.SpanExporter
	switch kind {
	case "newrelic":
		secure := strings.HasPrefix(cfg.NewRelicURL, "https")
		opts := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(cfg.NewRelicURL),
			otlptracegrpc.WithHeaders(map[string]string{
				"api-key": cfg.NewRelicKey,
			}),
		}
		if !secure {
			opts = append(opts, otlptracegrpc.WithInsecure())
		}
		driver := otlptracegrpc.NewClient(opts...)
		exp, err := otlptrace.New(context.Background(), driver)
		if err != nil {
			log.Errorf("creating OTLP trace exporter: %v", err)
			return nil, err
		}
		exporter = exp
		log.Info("using new relic exporter")
	case "stdout":
		otExporter, err := stdout.New(stdout.WithPrettyPrint())
		if err != nil {
			break
		}
		exporter = otExporter
		log.Info("using stdout exporter")
	case "jaeger":
		switch cfg.JaegerMode {
		case "collector":
			pr, err := oteljaeger.New(oteljaeger.WithCollectorEndpoint(oteljaeger.WithEndpoint(cfg.JaegerCollectorURL)))
			if err != nil {
				return nil, err
			}
			exporter = pr
			log.Info("using jaeger collector exporter")
		default:
			opts := []oteljaeger.AgentEndpointOption{
				oteljaeger.WithAgentHost(cfg.JaegerAgentHost),
				oteljaeger.WithAgentPort(cfg.JaegerAgentPort),
			}
			pr, err := oteljaeger.New(oteljaeger.WithAgentEndpoint(opts...))
			if err != nil {
				return nil, err
			}
			exporter = pr
			log.Info("using jaeger agent exporter")
		}
	case "otel":
		secure := strings.HasPrefix(cfg.OtelEndpoint, "https")
		var driver otlptrace.Client
		switch cfg.OtelMode {
		case "http":
			opts := []otlptracehttp.Option{otlptracehttp.WithEndpoint(cfg.OtelEndpoint)}
			if !secure {
				opts = append(opts, otlptracehttp.WithInsecure())
			}
			driver = otlptracehttp.NewClient(opts...)
			log.Info("using OTLP http exporter")
		case "grpc":
			opts := []otlptracegrpc.Option{otlptracegrpc.WithEndpoint(cfg.OtelEndpoint)}
			if !secure {
				opts = append(opts, otlptracegrpc.WithInsecure())
			}
			driver = otlptracegrpc.NewClient(opts...)
			log.Info("using OTLP GRPC exporter")
		default:
			return nil, errors.New("unsupported mode")
		}

		exp, err := otlptrace.New(context.Background(), driver)
		if err != nil {
			log.Errorf("creating OTLP trace exporter: %v", err)
			return nil, err
		}
		exporter = exp
	default:
		provider = trace.NewNoopTracerProvider()
		otel.SetTracerProvider(provider)
		return provider, nil
	}

	rate := float64(1)
	if cfg.SampleRate > 0 {
		rate = cfg.SampleRate
	}

	tp := sdk.NewTracerProvider(
		// Always be sure to batch in production.
		sdk.WithBatcher(exporter),
		// Record information about this application in an Resource.
		sdk.WithSampler(sdk.TraceIDRatioBased(rate)),
		sdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(name),
			attribute.String("environment", env.Get()),
			attribute.Int64("ID", 1),
		)),
	)

	otel.SetTracerProvider(tp)

	return tp, nil
}
