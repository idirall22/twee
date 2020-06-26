package tweet_test

import (
	"context"
	"io"
	"log"
	"net"
	"testing"
	"time"

	"github.com/idirall22/twee/follow"
	fpostgresstore "github.com/idirall22/twee/follow/store/postgres"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/idirall22/twee/auth"
	"github.com/idirall22/twee/common"
	sample "github.com/idirall22/twee/generator"
	"github.com/idirall22/twee/pb"
	"github.com/idirall22/twee/tweet"
	teventstore "github.com/idirall22/twee/tweet/event_store/stan"
	postgresstore "github.com/idirall22/twee/tweet/store/postgres"
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

	// start tweets server and get a tweets client
	followAddr := startFollowTestServer(t, jwtManager)
	followClient := startFollowClient(t, followAddr)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// Register new user 1.
	reqReg1 := sample.RandomRegisterRequest()
	res1, err := authClient.Register(ctx, reqReg1)
	require.NoError(t, err)
	require.NotNil(t, res1)

	// Login user and get jwt token.
	reqLogin1 := sample.LoginRequestFromRegisterRequest(reqReg1)
	resLogin1, err := authClient.Login(ctx, reqLogin1)
	require.NoError(t, err)
	require.NotNil(t, resLogin1)

	// Register new user 2.
	reqReg2 := sample.RandomRegisterRequest()
	res2, err := authClient.Register(ctx, reqReg2)
	require.NoError(t, err)
	require.NotNil(t, res2)

	// Login user and get jwt token.
	reqLogin2 := sample.LoginRequestFromRegisterRequest(reqReg2)
	resLogin2, err := authClient.Login(ctx, reqLogin2)
	require.NoError(t, err)
	require.NotNil(t, resLogin2)

	// adding jwt token to context.
	ctx1 := metadata.AppendToOutgoingContext(context.Background(), auth.AuthKey, resLogin1.AccessToken)
	userClaims1, err := jwtManager.Verify(resLogin1.AccessToken)
	require.NoError(t, err)
	require.NotNil(t, userClaims1)

	ctx2 := metadata.AppendToOutgoingContext(context.Background(), auth.AuthKey, resLogin2.AccessToken)

	resFollow, err := followClient.ToggleFollow(ctx2, &pb.RequestFollow{Followee: userClaims1.ID})
	require.NoError(t, err)
	require.NotNil(t, resFollow)

	// create tweets
	createdIds := []int64{}
	for i := 0; i < 2; i++ {
		reqCre := sample.NewRequestCreateTweet()
		resCreate, err := tweetClient.Create(ctx1, reqCre)
		require.NoError(t, err)
		require.NotNil(t, resCreate)
		createdIds = append(createdIds, resCreate.Id)
	}

	// List tweets
	reqList := sample.NewRequestListTweet(1)
	stream, err := tweetClient.List(ctx1, reqList)
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

// start auth server
func startFollowTestServer(t *testing.T, jwtManager *auth.JwtManager) string {
	pStore, err := fpostgresstore.NewPostgresFollowStore(common.PostgresTestOptions)
	require.NoError(t, err)
	require.NotNil(t, pStore)

	server, err := follow.NewFollowServer(pStore, nil)
	require.NoError(t, err)
	require.NotNil(t, server)

	jwtInterceptor := auth.NewJwtInterceptor(jwtManager)
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(jwtInterceptor.Unary()),
		grpc.StreamInterceptor(jwtInterceptor.Stream()),
	)
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
