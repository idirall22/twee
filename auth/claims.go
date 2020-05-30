package auth

import "github.com/dgrijalva/jwt-go"

// UserClaims user claims
type UserClaims struct {
	jwt.StandardClaims
	ID       int64
	Username string
}
