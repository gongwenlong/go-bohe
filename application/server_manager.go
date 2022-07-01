package application

import (
	"bohe/transport"
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"os"
	"os/signal"
	"syscall"
)

type SignalHandler func(app *ServerManager, sig os.Signal)

type ServerManager struct {
	serverList []transport.Server
	cancelFunc context.CancelFunc

	sigList       []os.Signal
	signalHandler SignalHandler
}

func NewServerManager() *ServerManager {
	return &ServerManager{
		sigList: make([]os.Signal, 0),
	}
}

func (s *ServerManager) AppendService(service transport.Server) {
	s.serverList = append(s.serverList, service)
}

func (s *ServerManager) Start(ctx context.Context) error {

	if s.cancelFunc != nil {
		return fmt.Errorf("server is already running")
	}

	if len(s.serverList) == 0 {
		return fmt.Errorf("application has no service to run")
	}

	ctx, cancelFunc := context.WithCancel(ctx)
	group, ctx := errgroup.WithContext(ctx)
	s.cancelFunc = cancelFunc

	for _, service := range s.serverList {
		localService := service
		group.Go(func() error {
			logrus.Infof("starting server:%s", localService.GetType())
			return localService.Start(ctx)
		})
	}

	if len(s.sigList) == 0 {
		return group.Wait()
	}

	c := make(chan os.Signal, len(s.sigList))
	signal.Notify(c, s.sigList...)
	group.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case sig := <-c:
				if s.signalHandler != nil {
					s.signalHandler(s, sig)
				}
			}
		}
	})
	return group.Wait()
}

func (s *ServerManager) Stop(ctx context.Context) error {
	if s.cancelFunc == nil {
		return fmt.Errorf("application is already stopping")
	}
	s.cancelFunc()

	for _, server := range s.serverList {
		_ = server.Stop(ctx)
	}

	s.cancelFunc = nil
	return nil
}

func (s *ServerManager) SetSignalHandler(handler SignalHandler) {
	s.sigList = []os.Signal{
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGINT,
	}
	s.signalHandler = func(s *ServerManager, sig os.Signal) {
		s.Stop(context.Background())
		handler(s, sig)
	}
}
