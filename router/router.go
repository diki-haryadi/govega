package router

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"runtime/debug"
	"strings"
	"time"

	"github.com/diki-haryadi/govega/monitor"
	"github.com/diki-haryadi/govega/response"
	"github.com/felixge/httpsnoop"
	"github.com/julienschmidt/httprouter"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/diki-haryadi/govega/log"
)

type (
	MyRouter struct {
		Httprouter     *httprouter.Router
		WrappedHandler http.Handler
		Options        *Options
		tracer         trace.Tracer
	}

	Options struct {
		Prefix  string
		Timeout int
	}

	captureConfig struct {
		captureHandler bool
	}

	httpParamsKey     struct{}
	responseWriterKey struct{}
	captureHandlerKey struct{}
)

var (
	// list of paths which we don't want to capture the data
	nonCapturePaths = []string{
		"/metrics",
		"/health",
		"/debug/pprof",
	}

	httpResponseLogKey = attribute.Key("http.response.log")
)

func GetHttpParam(ctx context.Context, name string) string {
	ps := ctx.Value(httpParamsKey{}).(httprouter.Params)
	return ps.ByName(name)
}

func GetResponseWriter(ctx context.Context) http.ResponseWriter {
	val := ctx.Value(responseWriterKey{})
	if val == nil {
		return nil
	}
	return val.(http.ResponseWriter)
}

func New(o *Options) *MyRouter {
	myrouter := &MyRouter{
		Options: o,
		tracer:  otel.Tracer("router/myrouter"),
	}
	myrouter.Httprouter = httprouter.New()
	return myrouter
}

type Handle func(*http.Request) *response.JSONResponse

func (mr *MyRouter) GET(path string, handle Handle) {
	mr.Handle(path, http.MethodGet, handle)
}

func (mr *MyRouter) POST(path string, handle Handle) {
	mr.Handle(path, http.MethodPost, handle)
}

func (mr *MyRouter) PUT(path string, handle Handle) {
	mr.Handle(path, http.MethodPut, handle)
}

func (mr *MyRouter) PATCH(path string, handle Handle) {
	mr.Handle(path, http.MethodPatch, handle)
}

func (mr *MyRouter) DELETE(path string, handle Handle) {
	mr.Handle(path, http.MethodDelete, handle)
}

func (mr *MyRouter) Handle(path, method string, handle Handle) {
	mr.registerHandle(path, method, handleToHttpRouterHandle(handle))
}

func (mr *MyRouter) Handler(path, method string, handler http.Handler) {
	mr.registerHandle(path, method, httpHandlerToHttpRouterHandle(handler))
}

func (mr *MyRouter) ServeFiles(path string, root http.FileSystem) {
	fullPath := mr.Options.Prefix + path
	mr.Httprouter.ServeFiles(fullPath, root)
}

func (mr *MyRouter) Group(path string, fn func(r *MyRouter)) {
	sr := &MyRouter{
		Options: &Options{
			Prefix:  mr.Options.Prefix + path,
			Timeout: mr.Options.Timeout,
		},
		Httprouter: mr.Httprouter,
		tracer:     mr.tracer,
	}
	fn(sr)
}

// Deprecated: use MyRouter instead, it's already implement http.Handler
func (mr *MyRouter) HttpHandler() http.Handler {
	return mr
}

func (mr *MyRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	captureConf := captureConfig{
		captureHandler: true,
	}
	ctx := context.WithValue(r.Context(), captureHandlerKey{}, &captureConf)

	m := httpsnoop.CaptureMetrics(mr.Httprouter, w, r.WithContext(ctx))

	if captureConf.captureHandler {
		monitor.FeedHTTPMetrics(m.Code, m.Duration, r.Header.Get("routePath"), r.Method)
	}
}

func (mr *MyRouter) registerHandle(path, method string, handle httprouter.Handle) {
	fullPath := mr.Options.Prefix + path
	log.Println(fullPath)
	mr.Httprouter.Handle(method, fullPath, mr.handlePath(fullPath, handle))
}

func (mr *MyRouter) handlePath(fullPath string, handle httprouter.Handle) httprouter.Handle {
	var captureHandler = true
	for _, path := range nonCapturePaths {
		if strings.Contains(fullPath, path) {
			captureHandler = false
			break
		}
	}

	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second*time.Duration(mr.Options.Timeout))
		defer cancel()

		ctx = context.WithValue(ctx, httpParamsKey{}, ps)

		if value, ok := ctx.Value(captureHandlerKey{}).(*captureConfig); ok {
			value.captureHandler = captureHandler
		}

		r.Header.Set("routePath", fullPath)

		r = r.WithContext(ctx)

		handle(w, r, ps)
	}
}

func httpHandlerToHttpRouterHandle(handler http.Handler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ctx := r.Context()

		ctx = context.WithValue(ctx, httpParamsKey{}, ps)

		panicHttpHandlerWrapper(handler).ServeHTTP(w, r.WithContext(ctx))
	}
}

func handleToHttpRouterHandle(handle Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		t := time.Now()

		ctx := r.Context()

		ctx = context.WithValue(ctx, httpParamsKey{}, ps)
		ctx = context.WithValue(ctx, responseWriterKey{}, w)

		respChan := make(chan *response.JSONResponse)

		r = r.WithContext(ctx)

		go func() {
			respChan <- panicHandleWrapper(handle)(r)
		}()

		select {
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				response.NewJSONResponse().SetError(response.ErrTimeoutError).Send(w)
			}
		case resp := <-respChan:
			if resp != nil {
				span := trace.SpanFromContext(ctx)
				span.SetAttributes(httpResponseLogKey.String(fmt.Sprintf("%+v", resp.Log)))

				resp.SetLatency(time.Since(t).Seconds() * 1000)

				if resp.StatusCode > 499 {
					m := map[string]interface{}{}
					httpDump := dumpRequest(r)
					m["ERROR:"] = resp.RealError
					m["RESPONSE:"] = string(resp.GetBody())
					m["DUMP:"] = string(httpDump)
					log.Printf("%+v", m)
				}

				resp.Send(w)
			} else {
				httpDump := dumpRequest(r)

				log.WithFields(log.Fields{
					"dump": string(httpDump),
				}).Errorln("[Router] Nil response received from the handler")

				response.NewJSONResponse().SetError(response.ErrInternalServerError).Send(w)
			}
		}
	}
}

func panicHttpHandlerWrapper(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				httpDump := dumpRequest(r)
				stackTrace := string(debug.Stack())

				log.WithFields(log.Fields{
					"path":       r.URL.Path,
					"httpDump":   string(httpDump),
					"stackTrace": stackTrace,
					"error":      fmt.Sprintf("%+v", err),
				}).Errorln("[router/panicHttpHandlerWrapper] panic have occurred")

				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}()

		handler.ServeHTTP(w, r)
	})
}

func panicHandleWrapper(handle Handle) Handle {
	return func(r *http.Request) (resp *response.JSONResponse) {
		defer func() {
			if err := recover(); err != nil {
				httpDump := dumpRequest(r)
				stackTrace := string(debug.Stack())

				log.WithFields(log.Fields{
					"path":       r.URL.Path,
					"httpDump":   string(httpDump),
					"stackTrace": stackTrace,
					"error":      fmt.Sprintf("%+v", err),
				}).Errorln("[router/panicHandleWrapper] panic have occurred")

				resp = response.NewJSONResponse().SetError(response.ErrInternalServerError)
				return
			}
		}()

		resp = handle(r)
		return
	}
}

func dumpRequest(r *http.Request) []byte {
	httpDump, err := httputil.DumpRequest(r, true)
	if err == nil {
		return httpDump
	}

	log.WithFields(log.Fields{
		"url":    r.URL,
		"method": r.Method,
		"header": fmt.Sprintf("%+v", r.Header),
		"err":    err,
	}).Debugln("[Router] Failed to dump request with body, re-attempting to dump request without body")

	// Retry without including body
	httpDump, err = httputil.DumpRequest(r, false)
	if err == nil {
		return httpDump
	}

	log.WithFields(log.Fields{
		"url":    r.URL,
		"method": r.Method,
		"header": fmt.Sprintf("%+v", r.Header),
		"err":    err,
	}).Infoln("[Router] Failed to dump request")
	return nil
}
