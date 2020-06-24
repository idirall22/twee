package tweet_test

import (
	"context"
	"io"
	"log"
	"net"
	"testing"
	"time"

	teventstore "github.com/idirall22/twee/tweet/event_store/stan"

	postgresstore "github.com/idirall22/twee/tweet/store/postgres"

	"google.golang.org/grpc/metadata"

	"github.com/idirall22/twee/auth"
	"github.com/idirall22/twee/common"

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
	tweetAddr := startTweetTestServer(t, jwtManager)
	tweetClient := startTweetClient(t, tweetAddr)

	// start tweets server and get a tweets client
	authAddr := startAuthTestServer(t, jwtManager)
	authClient := startAuthClient(t, authAddr)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// Register new user.
	reqReg := sample.RandomRegisterRequest()
	res, err := authClient.Register(ctx, reqReg)
	require.NoError(t, err)
	require.NotNil(t, res)

	// Login user and get jwt token.
	reqLogin := sample.LoginRequestFromRegisterRequest(reqReg)
	resLogin, err := authClient.Login(ctx, reqLogin)
	require.NoError(t, err)
	require.NotNil(t, resLogin)

	// adding jwt token to context.
	token := resLogin.AccessToken
	ctx = metadata.AppendToOutgoingContext(ctx, auth.AuthKey, token)

	// create tweets
	createdIds := []int64{}
	for i := 0; i < 2; i++ {
		reqCre := sample.NewRequestCreateTweet()
		resCreate, err := tweetClient.Create(ctx, reqCre)
		require.NoError(t, err)
		require.NotNil(t, resCreate)
		createdIds = append(createdIds, resCreate.Id)
	}

	// List tweets
	reqList := sample.NewRequestListTweet(1)
	stream, err := tweetClient.List(ctx, reqList)
	require.NoError(t, err)
	require.NotNil(t, stream)

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Error to receive stream of tweets: %v", err)
		}
		require.NotNil(t, res)
	}

	// // Delete tweets
	// for _, tweetId := range createdIds {
	// 	reqDel := sample.NewRequestDeleteTweet(tweetId)
	// 	_, err = tweetClient.Delete(ctx, reqDel)
	// 	require.NoError(t, err)
	// }

}

// start tweet server
func startTweetTestServer(t *testing.T, jwtManager *auth.JwtManager) string {
	pStore, err := postgresstore.NewPostgresTweetStore(common.PostgresTestOptions)
	require.NoError(t, err)
	require.NotNil(t, pStore)

	es, err := teventstore.NewNatsStreamingEventStore(
		"tweets",
		"test-cluster",
		"0111",
	)
	require.NoError(t, err)
	require.NotNil(t, pStore)

	server, err := tweet.NewTweetServer(pStore, es)
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
func startTweetClient(t *testing.T, address string) pb.TweetServiceClient {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	require.NoError(t, err)
	require.NotNil(t, conn)
	return pb.NewTweetServiceClient(conn)
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
