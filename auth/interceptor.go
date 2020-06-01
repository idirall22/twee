package auth

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"google.golang.org/grpc"
)

// ClaimKey used as a key context
type ClaimKey string

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
		claims, err := i.isAuthorized(ctx)
		if err != nil {
			return nil, err
		}

		ctx = context.WithValue(ctx, ClaimKey("claims"), claims)
		return handler(ctx, req)
	}
}

func (i *JwtInterceptor) isAuthorized(ctx context.Context) (*UserClaims, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "Unauthenticated")
	}

	values := md["authorization"]
	if len(values) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "Unauthenticated token not provided")
	}
	accessToken := values[0]

	claims, err := i.jwtManager.Verify(accessToken)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "access token not valid")
	}

	return claims, nil
}