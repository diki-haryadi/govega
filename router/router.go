package router

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"runtime/debug"
	"strings"
	"time"

	"github.com/felixge/httpsnoop"
	"github.com/julienschmidt/httprouter"

	"github.com/opentracing/opentracing-go"
	tlog "github.com/opentracing/opentracing-go/log"

	"gitlab.com/superman-tech/lib/log"
	"gitlab.com/superman-tech/lib/monitor"
	"gitlab.com/superman-tech/lib/response"
	customHttpUtil "gitlab.com/superman-tech/lib/util"
)

type MyRouter struct {
	Httprouter     *httprouter.Router
	WrappedHandler http.Handler
	Options        *Options
}

type Options struct {
	Prefix  string
	Timeout int
}

var (
	HttpRouter *httprouter.Router
)

func init() {
	HttpRouter = httprouter.New()
}

func WrapperHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := httpsnoop.CaptureMetrics(HttpRouter, w, r)
		if r.URL.String() != "/metrics" && !strings.Contains(r.URL.String(), "/healthz") { // Note: Dont want to monitor for /metrics and /healthz
			monitor.FeedHTTPMetrics(m.Code, m.Duration, r.Header.Get("routePath"), r.Method)
		}
	})
}

func New(o *Options) *MyRouter {
	myrouter := &MyRouter{Options: o}
	myrouter.Httprouter = HttpRouter
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
	fullPath := mr.Options.Prefix + path
	log.Println(fullPath)
	mr.Httprouter.Handle(method, fullPath, mr.handleNow(fullPath, handle))
}

func (mr *MyRouter) ServeFiles(path string, root http.FileSystem) {
	fullPath := mr.Options.Prefix + path
	mr.Httprouter.ServeFiles(fullPath, root)
}

func (mr *MyRouter) Group(path string, fn func(r *MyRouter)) {
	sr := New(&Options{
		Prefix:  mr.Options.Prefix + path,
		Timeout: mr.Options.Timeout,
	})
	fn(sr)
}

type panicObject struct {
	err        interface{}
	stackTrace string
}

func (mr *MyRouter) handleNow(fullPath string, handle Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		var (
			span opentracing.Span
			ctx  context.Context
		)

		t := time.Now()
		ctx, cancel := context.WithTimeout(r.Context(), time.Second*time.Duration(mr.Options.Timeout))

		defer cancel()

		ctx = context.WithValue(ctx, "HTTPParams", ps)
		ctx = context.WithValue(ctx, "ResponseWriter", w)

		if !strings.Contains(fullPath, "/healthz") {
			span, ctx = opentracing.StartSpanFromContext(ctx, r.RequestURI)
			defer span.Finish()
			span.LogFields(tlog.String("ip", customHttpUtil.GetClientIPAddress(r)))
		}

		r.Header.Set("routePath", fullPath)
		r = r.WithContext(ctx)

		respChan := make(chan *response.JSONResponse)
		recovered := make(chan panicObject)

		go func() {
			defer func() {
				if err := recover(); err != nil {
					recovered <- panicObject{
						err:        err,
						stackTrace: string(debug.Stack()),
					}
				}
			}()
			respChan <- handle(r)
		}()

		select {
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				response.NewJSONResponse().SetError(response.ErrTimeoutError).Send(w)
			}
		case cause := <-recovered:
			httpDump := dumpRequest(r)
			log.WithFields(log.Fields{
				"path":       r.URL.Path,
				"httpDump":   string(httpDump),
				"stackTrace": cause.stackTrace,
				"error":      fmt.Sprintf("%v", cause.err),
			}).Errorln("[Router] panic have occurred")
			response.NewJSONResponse().SetError(response.ErrInternalServerError).Send(w)
		case resp := <-respChan:
			if resp != nil {
				if span != nil {
					span.LogFields(tlog.Object("log", resp.Log))
					span.SetTag("httpCode", resp.StatusCode)
				}
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
				if fullPath != "/metrics" {
					httpDump := dumpRequest(r)
					log.WithFields(log.Fields{
						"dump": string(httpDump),
					}).Errorln("[Router] Nil response received from the handler")
					response.NewJSONResponse().SetError(response.ErrInternalServerError).Send(w)
				}
			}
		}
		return
	}
}

func GetHttpParam(ctx context.Context, name string) string {
	ps := ctx.Value("HTTPParams").(httprouter.Params)
	return ps.ByName(name)
}

func GetResponseWriter(ctx context.Context) http.ResponseWriter {
	val := ctx.Value("ResponseWriter")
	if val == nil {
		return nil
	}
	return val.(http.ResponseWriter)
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
	//Retry without including body
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
