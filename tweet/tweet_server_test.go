package tweet_test

import (
	"testing"

	"github.com/idirall22/twee/tweet"
	"github.com/stretchr/testify/require"
)

func TestServer(t *testing.T) {
	tweetServer, err := tweet.NewServer()
	require.NoError(t, err)
	require.NotNil(t, tweetServer)
	defer tweetServer.Close()
}
