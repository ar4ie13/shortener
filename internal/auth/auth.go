package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/ar4ie13/shortener/internal/service"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type Auth struct {
	Claims Claims
}

type Claims struct {
	jwt.RegisteredClaims
	UserUUID uuid.UUID
}

func NewAuth() *Auth {
	return &Auth{}
}

const (
	TokenExpiration = time.Hour * 24
	// SecretKey TODO: replace const with value from repository in real project
	SecretKey = "nHhjHgahbioHBGbBHJ"
)

func (a Auth) GenerateUserUUID() uuid.UUID {
	return uuid.New()
}

// BuildJWTString creates new JWT token
func (a Auth) BuildJWTString(userUUID uuid.UUID) (string, error) {
	// creating new token with HS256 algorithm and claims — Auth
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// token expiration date
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExpiration)),
		},
		// personal claim
		UserUUID: userUUID,
	})

	// creating signed token string
	tokenString, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

//// ValidateUserUUID validates token and return the UUID of user
//func (a Auth) ValidateUserUUID(tokenString string) (uuid.UUID, error) {
//	claims, token, err := a.parseTokenString(tokenString)
//	if err != nil {
//		if errors.Is(err, jwt.ErrTokenExpired) {
//			tokenString, err = a.BuildJWTString(claims.UserUUID)
//			if err != nil {
//				return uuid.Nil, err
//			}
//			claims, token, err = a.parseTokenString(tokenString)
//			if err != nil {
//				return uuid.Nil, err
//			}
//		} else {
//			return uuid.Nil, err
//		}
//	}
//	if claims.UserUUID.String() == "" {
//		return uuid.Nil, service.ErrInvalidUserUUID
//	}
//
//	if !token.Valid {
//		return uuid.Nil, fmt.Errorf("invalid token")
//	}
//	return claims.UserUUID, nil
//}

// ValidateUserUUID validates token and return the UUID of user
func (a Auth) ValidateUserUUID(tokenString string) (uuid.UUID, error) {
	claims, token, err := a.parseTokenString(tokenString)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			tokenString, err = a.BuildJWTString(claims.UserUUID)
			if err != nil {
				return uuid.Nil, err
			}
			claims, token, err = a.parseTokenString(tokenString)
			if err != nil {
				return uuid.Nil, err
			}
		} else {
			return uuid.Nil, err
		}
	}
	if claims.UserUUID.String() == "" || claims.UserUUID == uuid.Nil {
		return uuid.Nil, service.ErrInvalidUserUUID
	}

	if !token.Valid {
		return uuid.Nil, fmt.Errorf("invalid token")
	}
	return claims.UserUUID, nil
}

// parseTokenString parses token string and returns claims and token (for validation)
func (a Auth) parseTokenString(tokenString string) (*Claims, *jwt.Token, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(SecretKey), nil
	})
	if err != nil {
		return claims, token, err
	}
	return claims, token, nil
}
