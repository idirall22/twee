package sample

import (
	"math/rand"

	"github.com/brianvoe/gofakeit/data"
	"github.com/idirall22/twee/pb"
)

// NewRequestCreateTweet cerate new tweet
func NewRequestCreateTweet() *pb.CreateTweetRequest {
	return &pb.CreateTweetRequest{
		Content: randomWord(),
	}
}

// NewRequestUpdateTweet create new request update tweet
func NewRequestUpdateTweet(id int64) *pb.UpdateTweetRequest {
	return &pb.UpdateTweetRequest{
		Id:      id,
		Content: randomWord(),
	}
}

// NewRequestGetTweet create new request find tweet
func NewRequestGetTweet(id int64) *pb.GetTweetRequest {
	return &pb.GetTweetRequest{
		Id: id,
	}
}

// NewRequestListTweet create new request list tweet
func NewRequestListTweet(id int64) *pb.ListTweetRequest {
	return &pb.ListTweetRequest{}
}

// NewRequestDeleteTweet create new request delete tweet
func NewRequestDeleteTweet(id int64) *pb.DeleteTweetRequest {
	return &pb.DeleteTweetRequest{Id: id}
}

func randomWord() string {
	words := data.Lorem["word"]
	return words[rand.Intn(len(words))]
}
