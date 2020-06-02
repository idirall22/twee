package tweet_test

import (
	"context"
	"testing"
	"time"

	sample "github.com/idirall22/twee/generator"

	"github.com/idirall22/twee/tweet"
	"github.com/stretchr/testify/require"
)

func TestServer(t *testing.T) {
	tweetServer, err := tweet.NewServer()
	require.NoError(t, err)
	require.NotNil(t, tweetServer)

	defer tweetServer.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// create tweet
	reqCreate := sample.NewRequestCreateTweet()
	resCreate, err := tweetServer.Create(ctx, reqCreate)
	require.NoError(t, err)
	require.NotNil(t, resCreate)

	// update tweet
	reqUpdate := sample.NewRequestUpdateTweet(resCreate.Id)
	resUpdate, err := tweetServer.Update(ctx, reqUpdate)
	require.NoError(t, err)
	require.NotNil(t, resUpdate)

	// find tweet
	reqGet := sample.NewRequestGetTweet(resCreate.Id)
	resGet, err := tweetServer.Get(ctx, reqGet)
	require.NoError(t, err)
	require.NotNil(t, resGet)

	// delete tweet
	reqDel := sample.NewRequestDeleteTweet(resCreate.Id)
	resDel, err := tweetServer.Delete(ctx, reqDel)
	require.NoError(t, err)
	require.NotNil(t, resDel)
}
