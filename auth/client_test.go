package auth_test

import (
	"context"
	"net"
	"testing"
	"time"

	sample "github.com/idirall22/twee/generator"
	option "github.com/idirall22/twee/options"
	"github.com/idirall22/twee/pb"

	"github.com/idirall22/twee/auth"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestAuthService(t *testing.T) {
	addr := startAuthTestServer(t)
	client := startClient(t, addr)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// Register
	reqReg := sample.RandomRegisterRequest()
	resReg, err := client.Register(ctx, reqReg)
	require.NoError(t, err)
	require.NotNil(t, resReg)

	// Login
	reqLogin := sample.LoginRequestFromRegisterRequest(reqReg)
	resLog, err := client.Login(ctx, reqLogin)
	require.NoError(t, err)
	require.NotNil(t, resLog)
	require.NotEmpty(t, resLog.AccessToken)
}

// start auth server
func startAuthTestServer(t *testing.T) string {
	opts := option.NewPostgresOptions(
		"0.0.0.0",
		"postgres",
		"password",
		"twee",
		3,
		5432,
		time.Second,
	)
	jwtManager := auth.NewJwtManager(
		"secret",
		time.Minute*15,
		time.Hour*24*365,
	)
	server, err := auth.NewAuthServer(jwtManager, opts)
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
func startClient(t *testing.T, address string) pb.AuthServiceClient {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	require.NoError(t, err)
	require.NotNil(t, conn)
	return pb.NewAuthServiceClient(conn)
}
