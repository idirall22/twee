package timeline_test

import (
	"context"
	"io"
	"log"
	"math/rand"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/idirall22/twee/auth"
	"github.com/idirall22/twee/common"
	"github.com/idirall22/twee/follow"
	fpostgresstore "github.com/idirall22/twee/follow/store/postgres"
	sample "github.com/idirall22/twee/generator"
	"github.com/idirall22/twee/pb"
	"github.com/idirall22/twee/timeline"
	tlpostgresstore "github.com/idirall22/twee/timeline/store/postgres"
	"github.com/idirall22/twee/tweet"
	postgresstore "github.com/idirall22/twee/tweet/store/postgres"
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
	tmAddr := startTimelineTestServer(t, jwtManager, followClient)
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
		uctx := metadata.AppendToOutgoingContext(context.Background(), auth.AuthKey, accessTokens[i])

		for _, id := range f.followees {
			req := &pb.RequestFollow{Followee: id}
			res, err := followClient.ToggleFollow(uctx, req)
			require.NoError(t, err)
			require.NotNil(t, res)
		}

		for i := 0; i < 2; i++ {
			req := sample.NewRequestCreateTweet()
			res, err := tweetClient.Create(uctx, req)
			require.NoError(t, err)
			require.NotNil(t, res)
		}

	}

	uctx := metadata.AppendToOutgoingContext(
		context.Background(), auth.AuthKey, accessTokens[0])
	// userInfos, err := auth.GetUserInfosFromContext(uctx)
	// require.NoError(t, err)
	// require.NotNil(t, userInfos)

	stream, err := timelineClient.Timeline(
		uctx,
		&pb.TimelineRequest{
			Type:   pb.TimelineType_HOME,
			UserId: 1,
		})
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
		log.Println(resTimeline.Tweet.Id, resTimeline.Tweet.UserId)
	}
}

// start auth server
func startTimelineTestServer(
	t *testing.T,
	jwtManager *auth.JwtManager,
	fc pb.FollowServiceClient,
) string {

	pStore, err := tlpostgresstore.NewPostgresTimelineStore(common.PostgresTestOptions)
	require.NoError(t, err)
	require.NotNil(t, pStore)

	server, err := timeline.NewTimelineServer(pStore, nil, fc)
	require.NoError(t, err)
	require.NotNil(t, server)

	jwtInterceptor := auth.NewJwtInterceptor(jwtManager)
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(jwtInterceptor.Unary()),
		grpc.StreamInterceptor(jwtInterceptor.Stream()),
	)
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

// start tweet server
func startTweetTestServer(t *testing.T, jwtManager *auth.JwtManager) string {
	pStore, err := postgresstore.NewPostgresTweetStore(common.PostgresTestOptions)
	require.NoError(t, err)
	require.NotNil(t, pStore)

	server, err := tweet.NewTweetServer(pStore, nil)
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

// start tweet client
func startTweetClient(t *testing.T, address string) pb.TweetServiceClient {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	require.NoError(t, err)
	require.NotNil(t, conn)
	return pb.NewTweetServiceClient(conn)
}

func sstream() grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		log.Printf("--> stream interceptor: %s", method)

		return streamer(ctx, desc, cc, method, opts...)
	}
}
