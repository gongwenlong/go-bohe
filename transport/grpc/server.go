package grpc

import (
	"context"
	"google.golang.org/grpc"
	"net"
)

type Config struct {
	Addr string
	Name string
}

type Server struct {
	*grpc.Server
	config   Config
	listener net.Listener
}

func (s *Server) GetType() string {
	return "GrpcServer"
}

func (s *Server) Start(ctx context.Context) error {
	listener, err := net.Listen("tcp", s.config.Addr)
	if err != nil {
		return err
	}

	s.listener = listener

	return s.Serve(listener)
}

func (s *Server) Stop(ctx context.Context) error {
	s.GracefulStop()
	return nil
}

func NewGrpcServer(config Config, opt ...grpc.ServerOption) *Server {
	return &Server{
		config: config,
		Server: grpc.NewServer(opt...),
	}
}
