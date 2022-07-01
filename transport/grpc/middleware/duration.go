package middleware

import (
	"context"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"strconv"
	"time"
)

func DurationInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	now := time.Now().UnixNano()
	defer func() {
		cost := (float64(time.Now().UnixNano()) - float64(now)) / 1e6
		if cost > 100 {
			logrus.Warnln("handle", info.FullMethod, "too long ,using ", cost, "ms")
		} else {
			logrus.Infoln("handle", info.FullMethod, "using ", cost, "ms")
		}
	}()
	return handler(ctx, req)
}

// TimeoutInterceptor 超时控制
func TimeoutInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	//默认超时时间 2s，如果客户端传递了 metadata，则使用 metadata 里的 timeout_seconds
	timeoutSecond := 2.0
	var err error
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if len(md["timeout_seconds"]) > 0 {
			timeoutSecond, err = strconv.ParseFloat(md["timeout_seconds"][0], 64)
			if err != nil {
				timeoutSecond = 2.0
			}
		}
	}

	ctx2, cancel := context.WithTimeout(ctx, time.Duration(timeoutSecond*1000)*time.Millisecond)
	defer cancel()
	done := make(chan error, 1)
	response := make(chan interface{}, 1)
	panicChan := make(chan interface{}, 1)
	go func() {
		defer func() {
			if p := recover(); p != nil {
				panicChan <- p
			}
		}()
		res, err := handler(ctx2, req)
		response <- res
		done <- err
	}()

	select {
	case err := <-done:
		res := <-response
		return res, err
	case p := <-panicChan:
		panic(p)
	case <-ctx2.Done():
		return handler, errors.New("RPC timeout error")
	}
}
