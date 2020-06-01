package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/idirall22/twee/pb"

	"github.com/idirall22/twee/auth"
	"github.com/stretchr/testify/require"
)

func TestAuthServer(t *testing.T) {
	t.Parallel()
	server, err := auth.NewAuthServer()
	require.NoError(t, err)
	require.NotNil(t, server)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	reqRegister := &pb.RegisterRequest{
		Username: "user",
		Password: "password",
	}

	res, err := server.Register(ctx, reqRegister)
	require.NoError(t, err)
	require.Nil(t, res)

	reqLogin := &pb.LoginRequest{
		Username: "user",
		Password: "password",
	}

	resLogin, err := server.Login(ctx, reqLogin)
	require.NoError(t, err)
	require.NotNil(t, resLogin)
}
