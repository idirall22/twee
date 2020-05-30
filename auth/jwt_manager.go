package auth

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/idirall22/twee/pb"
)

// JwtManager jwt manager struct
type JwtManager struct {
	secret               string
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
}

// NewJwtManager create new jwt manager
func NewJwtManager(
	secret string,
	accessTokenDuration, refreshTokenDuration time.Duration,
) *JwtManager {
	return &JwtManager{
		secret:               secret,
		accessTokenDuration:  accessTokenDuration,
		refreshTokenDuration: refreshTokenDuration,
	}
}

// GenerateAccessToken generate access token.
func (j *JwtManager) GenerateAccessToken(user *pb.User) (string, error) {

	userClaims := &UserClaims{
		Username: user.Username,
		ID:       user.Id,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * j.accessTokenDuration).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, userClaims)

	return token.SignedString([]byte(j.secret))
}

// GenerateRefreshToken generate refresh token.
func (j *JwtManager) GenerateRefreshToken(user *pb.User) (string, error) {
	userClaims := &UserClaims{
		Username: user.Username,
		ID:       user.Id,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * j.refreshTokenDuration).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, userClaims)

	return token.SignedString([]byte(j.secret))
}

// Verify verify token
func (j *JwtManager) Verify(accessToken string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(
		accessToken,
		&UserClaims{},
		func(token *jwt.Token) (interface{}, error) {
			_, ok := token.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				return nil, fmt.Errorf("Unexpected signing method")
			}
			return []byte(j.secret), nil
		},
	)

	if err != nil {
		return nil, fmt.Errorf("Invalid token: %v", err)
	}

	claims, ok := token.Claims.(*UserClaims)

	if !ok {
		return nil, fmt.Errorf("Invalid claims: %v", err)
	}
	return claims, nil
}
