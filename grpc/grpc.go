package grpc

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/diki-haryadi/govega/log"
	"github.com/diki-haryadi/govega/monitor"
	"github.com/valyala/fasthttp/reuseport"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
)

type Options struct {
	Address                      string
	UnaryInterceptors            []grpc.UnaryServerInterceptor
	SkipRegisterReflectionServer bool
}

type GRPCServer struct {
	s       *grpc.Server
	address string
}

func New(o *Options) *GRPCServer {
	unaryServerInterceptors := append([]grpc.UnaryServerInterceptor{
		otelgrpc.UnaryServerInterceptor(),
		MetricsUnaryInterceptor,
		UnaryServerPanicInterceptor,
	}, o.UnaryInterceptors...)

	srv := grpc.NewServer(grpc.ChainUnaryInterceptor(unaryServerInterceptors...))

	if !o.SkipRegisterReflectionServer {
		reflection.Register(srv)
	}

	return &GRPCServer{
		s:       srv,
		address: o.Address,
	}
}

func (gs *GRPCServer) Register(fn interface{}, server interface{}) {
	v := reflect.ValueOf(fn)
	if v.Kind() != reflect.Func {
		log.Fatal("first parameter must be a function")
	}

	vargs := make([]reflect.Value, 2)
	vargs[0] = reflect.ValueOf(gs.s)
	vargs[1] = reflect.ValueOf(server)
	v.Call(vargs)
}

func (gs *GRPCServer) Start() error {
	l, err := reuseport.Listen("tcp4", gs.address)
	if err != nil {
		return err
	}

	log.Println("starting grpc server on ", gs.address)
	return gs.s.Serve(l)
}

func (gs *GRPCServer) MustStart() {
	err := gs.Start()
	if err != nil {
		log.Fatalf("failed to start grpc server: %v", err)
	}
}

func (gs *GRPCServer) CatchSignal(s os.Signal) {
	log.Println("grpc service got signal:", s)
	if s == syscall.SIGHUP || s == syscall.SIGTERM {
		gs.Stop()
	}
}

func (gs *GRPCServer) Stop() {
	gs.StopContext(context.Background())
}

func (gs *GRPCServer) StopContext(ctx context.Context) {
	done := make(chan bool)

	go func(doneCh chan<- bool) {
		gs.s.GracefulStop()
		doneCh <- true
	}(done)

	select {
	case <-ctx.Done():
		log.Warnln("timeout waiting grpc server to stopped...")
	case <-done:
		close(done)
		log.Infoln("grpc server stopped...")
	}
}

func MetricsUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()

	resp, err := handler(ctx, req)
	stat, _ := status.FromError(err)

	monitor.FeedGRPCMetrics(info.FullMethod, stat.Code(), time.Since(start))
	return resp, err
}

func UnaryServerPanicInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			var reqString string
			if req != nil {
				reqString = fmt.Sprintf("%+v", req)
			}

			log.WithFields(log.Fields{
				"err":         r,
				"full_method": info.FullMethod,
				"req":         reqString,
				"stacktrace":  string(stack),
			}).Errorln("[UnaryServerPanicInterceptor] panic on handling request")

			err = status.Error(codes.Internal, codes.Internal.String())
			return
		}
	}()

	return handler(ctx, req)
}
