package auth_test

import (
	"context"
	"testing"
	"time"

	sample "github.com/idirall22/twee/generator"

	"github.com/idirall22/twee/auth"
	"github.com/stretchr/testify/require"
)

func TestAuthServer(t *testing.T) {
	t.Parallel()

	// auth server
	server, err := auth.NewAuthServer(nil)
	require.NoError(t, err)
	require.NotNil(t, server)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// register new user
	reqRegister := sample.RandomRegisterRequest()
	res, err := server.Register(ctx, reqRegister)
	require.NoError(t, err)
	require.NotNil(t, res)

	// login user
	reqLogin := sample.LoginRequestFromRegisterRequest(reqRegister)
	resLogin, err := server.Login(ctx, reqLogin)
	require.NoError(t, err)
	require.NotNil(t, resLogin)
}
