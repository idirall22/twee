package follow_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/idirall22/twee/common"

	sample "github.com/idirall22/twee/generator"

	"github.com/idirall22/twee/auth"
	"github.com/idirall22/twee/follow"
	"github.com/idirall22/twee/pb"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestFollow(t *testing.T) {
	jwtManager := auth.NewJwtManager(
		"secret",
		time.Minute*15,
		time.Hour*24*365,
	)

	followAddr := startFollowTestServer(t, jwtManager)
	followClient := startFollowClient(t, followAddr)

	authAddr := startAuthTestServer(t, jwtManager)
	authClient := startAuthClient(t, authAddr)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// register user 1
	reqRegister := sample.RandomRegisterRequest()
	resRegister, err := authClient.Register(ctx, reqRegister)
	require.NoError(t, err)
	require.NotNil(t, resRegister)

	// register user 2
	reqRegister2 := sample.RandomRegisterRequest()
	resRegister2, err := authClient.Register(ctx, reqRegister2)
	require.NoError(t, err)
	require.NotNil(t, resRegister2)

	// Login user 1
	reqLogin := sample.LoginRequestFromRegisterRequest(reqRegister)
	resLogin, err := authClient.Login(ctx, reqLogin)
	require.NoError(t, err)
	require.NotNil(t, resLogin)

	// Login user 2
	reqLogin2 := sample.LoginRequestFromRegisterRequest(reqRegister2)
	resLogin2, err := authClient.Login(ctx, reqLogin2)
	require.NoError(t, err)
	require.NotNil(t, resLogin2)

	// adding auth token to context
	token := resLogin.AccessToken
	ctx = metadata.AppendToOutgoingContext(ctx, auth.AuthKey, token)

	userClaims, err := jwtManager.Verify(resLogin2.AccessToken)
	require.NoError(t, err)
	require.NotNil(t, userClaims)

	resFollow, err := followClient.ToggleFollow(ctx, &pb.RequestFollow{Followee: userClaims.ID})
	require.NoError(t, err)
	require.NotNil(t, resFollow)

	resFollow, err = followClient.ToggleFollow(ctx, &pb.RequestFollow{Followee: userClaims.ID})
	require.NoError(t, err)
	require.NotNil(t, resFollow)
}

// start auth server
func startFollowTestServer(t *testing.T, jwtManager *auth.JwtManager) string {
	server, err := follow.NewFollowServer()
	require.NoError(t, err)
	require.NotNil(t, server)

	jwtInterceptor := auth.NewJwtInterceptor(jwtManager)
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(jwtInterceptor.Unary()))
	pb.RegisterFollowServiceServer(grpcServer, server)

	listner, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	require.NotNil(t, listner)

	go grpcServer.Serve(listner)
	return listner.Addr().String()
}

// start auth client
func startFollowClient(t *testing.T, address string) pb.FollowServiceClient {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	require.NoError(t, err)
	require.NotNil(t, conn)
	return pb.NewFollowServiceClient(conn)
}

// start auth server
func startAuthTestServer(t *testing.T, jwtManager *auth.JwtManager) string {
	server, err := auth.NewAuthServer(jwtManager, common.PostgresTestOptions)
	require.NoError(t, err)
	require.NotNil(t, server)

	grpcServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcServer, server)

	listner, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	require.NotNil(t, listner)

	go grpcServer.Serve(listner)

	return listner.Addr().String()
}

// start auth client
func startAuthClient(t *testing.T, address string) pb.AuthServiceClient {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	require.NoError(t, err)
	require.NotNil(t, conn)
	return pb.NewAuthServiceClient(conn)
}
