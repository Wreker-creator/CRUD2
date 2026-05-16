package marketgrpc

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
)

func LoggingInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {

	start := time.Now()
	ret, err := handler(ctx, req)
	log.Printf("method: %s | duration: %v | error: %v", info.FullMethod, time.Since(start), err)
	return ret, err // always return what the handler gave you

}
