package auth

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	apstore "github.com/idirall22/twee/auth/store/postgres"
	option "github.com/idirall22/twee/options"
	"github.com/idirall22/twee/pb"
)

// Server auth server struct
type Server struct {
	authStore *apstore.PostgresAuthStore
}

// NewAuthServer create new auth store
func NewAuthServer() (*Server, error) {
	opts := option.NewPostgresOptions(
		"0.0.0.0",
		"postgres",
		"password",
		"auth_tweets",
		3,
		5432,
		time.Second,
	)

	aStore, err := apstore.NewPostgresAuthStore(opts)

	if err != nil {
		return nil, fmt.Errorf("Could not Start store: %v", err)
	}
	return &Server{
		authStore: aStore,
	}, nil
}

// Register new user
func (s *Server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if len(req.Username) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Empty Username")
	}

	if len(req.Password) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Empty Password")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.GetPassword()), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not hash password: %v", err)
	}

	err = s.authStore.Create(ctx, req.GetUsername(), string(hashedPassword))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not Create user")
	}

	return nil, nil
}

// Login authenticate user
func (s *Server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	username := req.GetUsername()
	password := req.GetPassword()

	if len(username) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Empty Username")
	}

	if len(password) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Empty Password")
	}

	user, err := s.authStore.Find(ctx, username)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not Find user: %v", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.GetHashPassword()), []byte(password))
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Password not valid")
	}

	res := &pb.LoginResponse{
		AccessToken:  "",
		RefreshToken: "",
	}

	return res, nil
}

// Logout user
func (s *Server) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	return nil, nil
}
