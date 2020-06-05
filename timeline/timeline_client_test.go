package timeline_test

import (
	"context"
	"io"
	"log"
	"math/rand"
	"net"
	"testing"
	"time"

	sample "github.com/idirall22/twee/generator"

	"github.com/idirall22/twee/tweet"

	"github.com/idirall22/twee/follow"
	"github.com/idirall22/twee/timeline"

	"github.com/idirall22/twee/auth"
	"github.com/idirall22/twee/pb"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func init() {
	rand.Seed(time.Now().Unix())
}
func TestTimeline(t *testing.T) {
	t.Parallel()

	jwtManager := auth.NewJwtManager(
		"secret",
		time.Minute*15,
		time.Hour*24*365,
	)
	// starting auth server
	authAddr := startAuthTestServer(t, jwtManager)
	authClient := startAuthClient(t, authAddr)

	// starting follow server
	fAddr := startFollowTestServer(t, jwtManager)
	followClient := startFollowClient(t, fAddr)

	// starting tweet server
	tAddr := startTweetTestServer(t, jwtManager)
	tweetClient := startTweetClient(t, tAddr)

	// starting timeline server
	tmAddr := startTimelineTestServer(t, jwtManager)
	timelineClient := startTimelineClient(t, tmAddr)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	accessTokens := []string{}
	for i := 0; i < 5; i++ {
		req := sample.RandomRegisterRequest()
		res, err := authClient.Register(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, res)

		reqLogin := sample.LoginRequestFromRegisterRequest(req)
		resLogin, err := authClient.Login(ctx, reqLogin)
		require.NoError(t, err)
		require.NotNil(t, resLogin)

		accessTokens = append(accessTokens, resLogin.AccessToken)
	}

	follow := []struct{ followees []int64 }{
		{followees: []int64{2, 3, 4, 5}},
		{followees: []int64{1, 4, 5}},
		{followees: []int64{1, 2, 5}},
		{followees: []int64{2, 3, 5}},
		{followees: []int64{1, 2, 3, 4}},
	}

	for i, f := range follow {
		token := accessTokens[i]
		ctx = metadata.AppendToOutgoingContext(ctx, auth.AuthKey, token)

		for _, id := range f.followees {
			req := &pb.RequestFollow{Followee: id}
			res, err := followClient.ToggleFollow(ctx, req)
			require.NoError(t, err)
			require.NotNil(t, res)
		}

		for i := 0; i < 5; i++ {
			req := sample.NewRequestCreateTweet()
			res, err := tweetClient.Create(ctx, req)
			require.NoError(t, err)
			require.NotNil(t, res)
		}

	}
	ctx = metadata.AppendToOutgoingContext(ctx, auth.AuthKey, accessTokens[0])

	stream, err := timelineClient.Timeline(ctx, &pb.TimelineRequest{Type: true})
	require.NoError(t, err)
	require.NotNil(t, stream)

	for {
		resTimeline, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Error to fetch user timeline: %v", err)
		}
		require.NotNil(t, resTimeline)
		log.Println(resTimeline.Tweet.Id)
	}
}

// start auth server
func startTimelineTestServer(t *testing.T, jwtManager *auth.JwtManager) string {
	server, err := timeline.NewTimelineServer()
	require.NoError(t, err)
	require.NotNil(t, server)

	jwtInterceptor := auth.NewJwtInterceptor(jwtManager)
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(jwtInterceptor.Unary()))
	pb.RegisterTimelineServiceServer(grpcServer, server)

	listner, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	require.NotNil(t, listner)

	go grpcServer.Serve(listner)
	return listner.Addr().String()
}

// start auth client
func startTimelineClient(t *testing.T, address string) pb.TimelineServiceClient {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	require.NoError(t, err)
	require.NotNil(t, conn)
	return pb.NewTimelineServiceClient(conn)
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

// start tweet server
func startTweetTestServer(t *testing.T, jwtManager *auth.JwtManager) string {
	server, err := tweet.NewServer()
	require.NoError(t, err)
	require.NotNil(t, server)

	grpcServer := grpc.NewServer()
	pb.RegisterTweetServiceServer(grpcServer, server)

	listner, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	require.NotNil(t, listner)

	go grpcServer.Serve(listner)

	return listner.Addr().String()
}

// start tweet client
func startTweetClient(t *testing.T, address string) pb.TweetServiceClient {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	require.NoError(t, err)
	require.NotNil(t, conn)
	return pb.NewTweetServiceClient(conn)
}
