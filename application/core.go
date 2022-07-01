package application

import (
	"context"
	transportGrpc "github.com/gongwenlong/go-bohe/transport/grpc"
	"github.com/gongwenlong/go-bohe/transport/grpc/middleware"
	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcRecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/sirupsen/logrus"
	"go.elastic.co/apm/module/apmgrpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
)

type CoreMicroServiceConfig struct {
	LogLevel string               `yaml:",omitempty"`
	Grpc     transportGrpc.Config `yaml:",omitempty"`
}

type CoreMicroService struct {
	*ServerManager

	config CoreMicroServiceConfig
}

func NewCoreMicroService() *CoreMicroService {
	return &CoreMicroService{
		ServerManager: NewServerManager(),
	}
}

func (c *CoreMicroService) InitGrpcServer(ctx context.Context, process func(grpcServer *transportGrpc.Server) error) error {
	if c.config.Grpc.Addr == "" {
		c.config.Grpc.Addr = "127.0.0.1:7881"
		//return fmt.Errorf("no grpc config set")
	}

	recoverFunc := func(p interface{}) (err error) {
		log.Println("error handle request: ", p)
		return status.Errorf(codes.Internal, "server error")
	}
	opts := []grpcRecovery.Option{
		grpcRecovery.WithRecoveryHandler(recoverFunc),
	}

	grpcServer := transportGrpc.NewGrpcServer(
		c.config.Grpc,
		grpcMiddleware.WithUnaryServerChain(
			grpcRecovery.UnaryServerInterceptor(opts...),
			apmgrpc.NewUnaryServerInterceptor(apmgrpc.WithRecovery()),
			middleware.DurationInterceptor,
			middleware.TimeoutInterceptor,
		),
		grpcMiddleware.WithStreamServerChain(
			grpcRecovery.StreamServerInterceptor(opts...),
		),
	)

	if err := process(grpcServer); err != nil {
		return err
	}

	c.ServerManager.AppendService(grpcServer)

	logrus.WithContext(ctx).Infof("[CoreMicroService] grpc init")
	return nil
}
