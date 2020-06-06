package user_test

import (
	"context"
	"io"
	"net"
	"testing"
	"time"

	sample "github.com/idirall22/twee/generator"

	"github.com/idirall22/twee/auth"

	option "github.com/idirall22/twee/options"
	"github.com/idirall22/twee/user"

	"github.com/idirall22/twee/pb"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestUserService(t *testing.T) {
	jwtManager := auth.NewJwtManager(
		"secret",
		time.Minute*15,
		time.Hour*24*365,
	)

	addr := startUserTestServer(t)
	userClient := startUserClient(t, addr)

	authAddr := startAuthTestServer(t, jwtManager)
	authClient := startAuthClient(t, authAddr)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	username := ""
	for i := 0; i < 10; i++ {
		req := sample.RandomRegisterRequest()
		username = req.Username
		res, err := authClient.Register(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, res)
	}

	// list users
	stream, err := userClient.List(ctx, &pb.RequestListUsers{Limit: 10, Offset: 0})
	require.NoError(t, err)
	require.NotNil(t, stream)

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		require.NotEmpty(t, res.User.Username)
	}

	// Get single user
	res, err := userClient.Profile(ctx, &pb.RequestUserProfile{Username: username})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.NotEmpty(t, res.User.Username)
}

// start auth server
func startUserTestServer(t *testing.T) string {
	opts := option.NewPostgresOptions(
		"0.0.0.0",
		"postgres",
		"password",
		"twee",
		3,
		5432,
		time.Second,
	)

	server, err := user.NewUserServer(opts)
	require.NoError(t, err)
	require.NotNil(t, server)

	grpcServer := grpc.NewServer()
	pb.RegisterUserServiceServer(grpcServer, server)
	listner, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	require.NotNil(t, listner)

	go grpcServer.Serve(listner)
	return listner.Addr().String()
}

// start auth client
func startUserClient(t *testing.T, address string) pb.UserServiceClient {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	require.NoError(t, err)
	require.NotNil(t, conn)
	return pb.NewUserServiceClient(conn)
}

// start auth server
func startAuthTestServer(t *testing.T, jwtManager *auth.JwtManager) string {
	server, err := auth.NewAuthServer(jwtManager)
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
