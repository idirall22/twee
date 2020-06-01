package sample

import (
	"math/rand"
	"time"

	"github.com/idirall22/twee/pb"

	fake "github.com/brianvoe/gofakeit/data"
)

func init() {
	rand.Seed(time.Now().Unix())
}

// RandomRegisterRequest create random auth register request
func RandomRegisterRequest() *pb.RegisterRequest {
	persons := fake.Person["first"]
	return &pb.RegisterRequest{
		Username: persons[rand.Intn(len(persons))],
		Password: "password",
	}
}

// LoginRequestFromRegisterRequest convert register request to login request
func LoginRequestFromRegisterRequest(req *pb.RegisterRequest) *pb.LoginRequest {
	return &pb.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	}
}
