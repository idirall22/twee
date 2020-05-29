package memorystore

import (
	"context"
	"sync"

	"github.com/google/uuid"

	"github.com/idirall22/twee/pb"
	"github.com/idirall22/twee/utils"
)

// TweetMemoryStore a memory store used only for dev
type TweetMemoryStore struct {
	mutex  sync.RWMutex
	tweets map[string]map[string]*pb.Tweet
}

// NewTweetMemoryStore create new memory tweet
func NewTweetMemoryStore() *TweetMemoryStore {
	return &TweetMemoryStore{
		tweets: make(map[string]map[string]*pb.Tweet),
	}
}

// Create tweet and return id
func (t *TweetMemoryStore) Create(ctx context.Context, tweet *pb.Tweet) (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", utils.ErrUUID
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	tweet.Id = id.String()
	userTweets := t.tweets[tweet.UserId]

	// check if user has tweets
	if userTweets == nil {
		t.tweets[tweet.UserId] = make(map[string]*pb.Tweet)
		userTweets = t.tweets[tweet.UserId]
	}

	userTweets[id.String()] = tweet

	return id.String(), nil
}

// Update update a tweet
func (t *TweetMemoryStore) Update(ctx context.Context, tweet *pb.Tweet) error {
	if len(tweet.Id) == 0 {
		return utils.ErrInvalidID
	}

	if len(tweet.UserId) == 0 {
		return utils.ErrInvalidUserID
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	userTweets := t.tweets[tweet.UserId]
	if userTweets == nil {
		return utils.ErrUpdate
	}

	mTweet := userTweets[tweet.Id]
	if mTweet == nil {
		return utils.ErrNotExists
	}

	mTweet.Content = tweet.Content

	return nil
}

// Delete tweet
func (t *TweetMemoryStore) Delete(ctx context.Context, id, userID string) error {
	if len(id) == 0 {
		return utils.ErrInvalidID
	}

	if len(userID) == 0 {
		return utils.ErrInvalidUserID
	}

	t.mutex.RLock()
	defer t.mutex.RUnlock()

	userTweets := t.tweets[userID]
	if userTweets == nil {
		return utils.ErrNotExists
	}
	_, ok := userTweets[id]
	if ok {
		delete(userTweets, id)
	}

	return nil
}

// Get tweet
func (t *TweetMemoryStore) Get(ctx context.Context, id, userID string) (*pb.Tweet, error) {
	if len(id) == 0 {
		return nil, utils.ErrInvalidID
	}

	if len(userID) == 0 {
		return nil, utils.ErrInvalidUserID
	}

	t.mutex.RLock()
	defer t.mutex.RUnlock()

	userTweets := t.tweets[userID]
	if userTweets == nil {
		return nil, utils.ErrUserRecordNotExists
	}

	tweet, ok := userTweets[id]
	if !ok {
		return nil, utils.ErrInvalidID
	}

	return tweet, nil
}

// List tweets
func (t *TweetMemoryStore) List(ctx context.Context, userID string, found func(tweet *pb.Tweet) error) error {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	userTweets, ok := t.tweets[userID]
	if !ok {
		return utils.ErrNotExists
	}

	for _, tweet := range userTweets {
		err := found(tweet)
		if err != nil {
			return err
		}
	}
	return nil
}
