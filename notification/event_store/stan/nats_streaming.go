package seventstore

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/idirall22/twee/follow"

	"github.com/nats-io/stan.go"
	// postgres driver
	_ "github.com/lib/pq"

	"github.com/idirall22/twee/common"
	"github.com/idirall22/twee/notification/store"
	"github.com/idirall22/twee/pb"
)

// NatsStreamingEventStore struct.
type NatsStreamingEventStore struct {
	notificationStore  store.Store
	followService      *follow.Server
	cc                 stan.Conn
	subject            string
	tweetNotifications chan string
	notifications      chan *pb.Notification
	done               chan error
}

// NewNatsStreamingEventStore create new NatsStreamingEventStore.
func NewNatsStreamingEventStore(
	subject, clusterID, clientID string,
	ns store.Store,
	fs *follow.Server,
) (*NatsStreamingEventStore, error) {

	cc, err := stan.Connect(clusterID, clientID)
	if err != nil {
		return nil, fmt.Errorf("Could not connect to nats streaming server: %v", err)
	}

	return &NatsStreamingEventStore{
		notificationStore:  ns,
		followService:      fs,
		subject:            subject,
		cc:                 cc,
		tweetNotifications: make(chan string, 128),
	}, nil
}

// Start NatsStreamingEventStore.
func (e *NatsStreamingEventStore) Start() error {
	log.Println("Event store started")

	go func() {
		for {
			msg := <-e.tweetNotifications
			tn := &pb.TweetEvent{}
			err := common.JSONToProtobufMessage(msg, tn)
			if err != nil {
				log.Println("----------------Error JSON")
				e.done <- fmt.Errorf("Could not parse json: %v", err)
				return
			}

			var followersList []*pb.Follow
			// ctx := metadata.AppendToOutgoingContext(context.Background(), auth.AuthKey, uc.Token)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			res, err := e.followService.ListFollow(ctx, &pb.RequestListFollow{
				FollowType: pb.FollowListType_FOLLOWEE,
				Followee:   tn.UserId,
			})

			if err != nil {
				log.Println("----------------Error ListFollows", err.Error())
				e.done <- fmt.Errorf("Could not list followers: %v", err)
				break
			}

			followersList = res.Follows

			if len(followersList) == 0 {
				log.Println("User have no followers yet")
				continue
			}
			err = e.notificationStore.NewTweetNotification(ctx, followersList, tn, e.notifications)
			if err != nil {
				e.done <- fmt.Errorf("Could not create notifications: %v", err)
				break
			}
		}
	}()

	sub, err := e.cc.Subscribe(e.subject, func(msg *stan.Msg) {
		log.Println("New notification received")
		log.Println(string(msg.Data))
		e.tweetNotifications <- string(msg.Data)
	})

	if err != nil {
		return fmt.Errorf("Could not subscribe to nats: %v", err)
	}

	defer sub.Close()
	return <-e.done
}

// Close close
func (e *NatsStreamingEventStore) Close() error {
	return e.cc.Close()
}

// Subscribe get notification channel
func (e *NatsStreamingEventStore) Subscribe() <-chan *pb.Notification {
	return e.notifications
}
