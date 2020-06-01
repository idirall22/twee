package auth

import (
	"context"
	"log"

	"google.golang.org/grpc"
)

// JwtInterceptor struct
type JwtInterceptor struct {
	jwtManager *JwtManager
}

// NewJwtInterceptor create new auth interceptor
func NewJwtInterceptor(jwtManager *JwtManager) *JwtInterceptor {
	return &JwtInterceptor{
		jwtManager: jwtManager,
	}
}

// Unary check if there is a token in the request context
func (i *JwtInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		log.Println("---------------------UNARY")
		return handler(ctx, req)
	}
}
