package auth_test

import (
	"context"
	"log"
	"net"
	"testing"
	"time"

	sample "github.com/idirall22/twee/generator"
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

	log.Println("Server running on: ", addr)
	reqReg := sample.RandomRegisterRequest()
	resReg, err := client.Register(ctx, reqReg)
	require.NoError(t, err)
	require.NotNil(t, resReg)

	reqLogin := sample.LoginRequestFromRegisterRequest(reqReg)
	resLog, err := client.Login(ctx, reqLogin)
	require.NoError(t, err)
	require.NotNil(t, resLog)
}

// start auth server
func startAuthTestServer(t *testing.T) string {
	server, err := auth.NewAuthServer()
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
