package auth

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
)

// ClaimKey used as a key context
type ClaimKey string

// AuthKey used as key for access token
var AuthKey = "authorization"

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

// Stream interceptor
func (i *JwtInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		ctx := ss.Context()
		claims, err := i.isAuthorized(ctx)
		if err != nil {
			return err
		}
		newStream := grpc_middleware.WrapServerStream(ss)
		newStream.WrappedContext = context.WithValue(ctx, ClaimKey("claims"), claims)
		return handler(srv, newStream)
	}
}

// GetUserInfosFromContext get user claims from context
func GetUserInfosFromContext(ctx context.Context) (*UserClaims, error) {
	userInfos, ok := ctx.Value(ClaimKey("claims")).(*UserClaims)
	if !ok {
		return nil, fmt.Errorf("could not parse user claim data")
	}
	return userInfos, nil
}

func (i *JwtInterceptor) isAuthorized(ctx context.Context) (*UserClaims, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "Unauthenticated")
	}
	values := md[AuthKey]
	if len(values) == 0 {
		userInfos, ok := ctx.Value(ClaimKey("claims")).(*UserClaims)
		fmt.Println(userInfos, ok)
		if ok {
			return userInfos, nil
		}
		return nil, status.Errorf(codes.Unauthenticated, "Unauthenticated token not provided")
	}

	accessToken := values[0]
	claims, err := i.jwtManager.Verify(accessToken)
	claims.Token = accessToken

	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "access token not valid")
	}

	return claims, nil
}
