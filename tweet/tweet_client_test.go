package tweet_test

import (
	"context"
	"io"
	"log"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc/metadata"

	"github.com/idirall22/twee/auth"

	sample "github.com/idirall22/twee/generator"
	"github.com/idirall22/twee/pb"
	"github.com/idirall22/twee/tweet"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestCreateTweets(t *testing.T) {
	t.Parallel()

	jwtManager := auth.NewJwtManager(
		"secret",
		time.Minute*15,
		time.Hour*24*365,
	)

	// start tweets server and get a tweets client
	addr := startAuthTestServer(t, jwtManager)
	client := startClient(t, addr)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	authServer, err := auth.NewAuthServer(jwtManager)
	require.NoError(t, err)
	require.NotNil(t, authServer)

	reqReg := sample.RandomRegisterRequest()
	_, err = authServer.Register(ctx, reqReg)
	require.NoError(t, err)
	resLogin, err := authServer.Login(ctx, &pb.LoginRequest{
		Username: reqReg.Username,
		Password: reqReg.Password,
	})

	token := resLogin.AccessToken
	ctx = metadata.AppendToOutgoingContext(ctx, auth.AuthKey, token)

	// create tweets
	createdIds := []int64{}
	for i := 0; i < 10; i++ {
		reqCre := sample.NewRequestCreateTweet()
		resCreate, err := client.Create(ctx, reqCre)
		require.NoError(t, err)
		require.NotNil(t, resCreate)
		createdIds = append(createdIds, resCreate.Id)
	}

	// List tweets
	reqList := sample.NewRequestListTweet(1)
	stream, err := client.List(ctx, reqList)
	require.NoError(t, err)
	require.NotNil(t, stream)

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Error to receive tweets: %v", err)
		}
		require.NotNil(t, res)
	}

	// Delete tweets
	for _, tweetId := range createdIds {
		reqDel := sample.NewRequestDeleteTweet(tweetId)
		_, err = client.Delete(ctx, reqDel)
		require.NoError(t, err)
	}
}

// start auth server
func startAuthTestServer(t *testing.T, jwtManager *auth.JwtManager) string {
	server, err := tweet.NewServer()
	require.NoError(t, err)
	require.NotNil(t, server)

	jwtInterceptor := auth.NewJwtInterceptor(jwtManager)
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(jwtInterceptor.Unary()))
	pb.RegisterTweetServiceServer(grpcServer, server)

	listner, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	require.NotNil(t, listner)

	go grpcServer.Serve(listner)
	return listner.Addr().String()
}

// start auth client
func startClient(t *testing.T, address string) pb.TweetServiceClient {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	require.NoError(t, err)
	require.NotNil(t, conn)
	return pb.NewTweetServiceClient(conn)
}
