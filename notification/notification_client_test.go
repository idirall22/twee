package notification_test

import (
	"context"
	"io"
	"log"
	"net"
	"testing"
	"time"

	"github.com/idirall22/twee/follow"
	sample "github.com/idirall22/twee/generator"

	"github.com/idirall22/twee/notification"

	"github.com/idirall22/twee/auth"
	"github.com/idirall22/twee/common"
	"github.com/idirall22/twee/pb"
	"github.com/idirall22/twee/tweet"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestNotificationSerivce(t *testing.T) {
	t.Parallel()

	jwtManager := auth.NewJwtManager(
		"secret",
		time.Minute*15,
		time.Hour*24*365,
	)

	authAddr := startAuthTestServer(t, jwtManager)
	authClient := startAuthClient(t, authAddr)

	tweetAddr := startTweetTestServer(t, jwtManager)
	tweetClient := startTweetClient(t, tweetAddr)

	followAddr := startFollowTestServer(t, jwtManager)
	followClient := startFollowClient(t, followAddr)

	notifAddr := startNotificationTestServer(t, jwtManager)
	notifClient := startNotitficationClient(t, notifAddr)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	stream, err := notifClient.Notify(ctx, &pb.NotifyRequest{})
	require.NoError(t, err)
	require.NotNil(t, stream)

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

	// Register new user.
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
	token2 := resLogin2.AccessToken
	ctx = metadata.AppendToOutgoingContext(ctx, auth.AuthKey, token2)

	userClaim1, err := jwtManager.Verify(token)
	require.NoError(t, err)
	require.NotNil(t, userClaim1)

	userClaim2, err := jwtManager.Verify(token2)
	require.NoError(t, err)
	require.NotNil(t, userClaim2)

	// user 2 follow user 1
	resFollow, err := followClient.ToggleFollow(ctx, &pb.RequestFollow{Followee: userClaim1.ID})
	require.NoError(t, err)
	require.NotNil(t, resFollow)

	ctx = metadata.AppendToOutgoingContext(context.Background(), auth.AuthKey, token)

	// create tweets
	createdIds := []int64{}
	for i := 0; i < 1; i++ {
		reqCre := sample.NewRequestCreateTweet()
		resCreate, err := tweetClient.Create(ctx, reqCre)
		require.NoError(t, err)
		require.NotNil(t, resCreate)
		createdIds = append(createdIds, resCreate.Id)
	}

	go func() {
		for {
			time.Sleep(1)
			n, err := stream.Recv()
			if err == io.EOF {
				log.Println("no more data")
				break
			}
			if err != nil {
				log.Fatal(err)
			}
			log.Println("|||||||||||||||||||||||||||||||||||||||")
			log.Println("New Notif: ", n.Notification.Title)
			log.Println("|||||||||||||||||||||||||||||||||||||||")
		}
	}()
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
	server, err := tweet.NewTweetServer(common.PostgresTestOptions)
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

// start Notitfication server
func startNotificationTestServer(t *testing.T, jwtManager *auth.JwtManager) string {
	server, err := notification.NewNotificationServer(common.PostgresTestOptions)
	require.NoError(t, err)
	require.NotNil(t, server)

	grpcServer := grpc.NewServer()
	pb.RegisterNotificationServiceServer(grpcServer, server)

	listner, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	require.NotNil(t, listner)

	go grpcServer.Serve(listner)
	return listner.Addr().String()
}

// start Notitfication client
func startNotitficationClient(t *testing.T, address string) pb.NotificationServiceClient {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	require.NoError(t, err)
	require.NotNil(t, conn)
	return pb.NewNotificationServiceClient(conn)
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
