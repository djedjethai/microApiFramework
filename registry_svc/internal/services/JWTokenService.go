package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	// "github.com/dgrijalva/jwt-go"
	"github.com/golang-jwt/jwt"
	e "gitlab.com/grpasr/common/errors/json"
	// "time"
)

// TODO pass these const via var env(secrets)
const (
	signedKey = "mySecretKey"
)

// JWTAccessGenerate generate the jwt access token
type JWTokenService struct {
	signedKey []byte // secret key
}

func NewJWTokenService() *JWTokenService {
	return &JWTokenService{
		signedKey: []byte(signedKey),
	}
}

// JWTokenIsValidToken valid the jwt_token, if valid return nil
func (a *JWTokenService) JWTokenIsValidToken(ctx context.Context, tokenString string) (map[string]string, e.IError) {

	var secretKey = a.signedKey
	tokenString = strings.TrimSpace(tokenString)
	tokenString = strings.Trim(tokenString, `"`)

	// fmt.Println("tokenString ---: ", tokenString)
	var claims jwt.MapClaims
	var ok bool

	_, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {

		// claims := token.Claims.(jwt.MapClaims)
		claims, ok = token.Claims.(jwt.MapClaims)
		if !ok {
			fmt.Println("err1")
			return nil, errors.New("ErrInvalidJWToken")
		}

		err := claims.Valid()
		if err != nil {
			if err.Error() == "Token is expired" {
				return nil, errors.New("ErrExpiredJWToken")
			}

			return nil, errors.New("ErrInvalidJWToken")
		}

		return secretKey, nil
	})
	if err != nil {
		if err.Error() == "ErrExpiredJWToken" {
			return nil, e.NewCustomHTTPStatus(e.StatusUnauthorized)
		}
		return nil, e.NewCustomHTTPStatus(e.StatusForbidden)
	}

	// Token is valid process it
	// Extracting the nested "openidInfo" map
	openidInfo, ok := claims["openidInfo"].(map[string]interface{})
	if !ok {
		return nil, e.NewCustomHTTPStatus(e.StatusForbidden)
	}

	infos := make(map[string]string)

	// Extracting role, sub, and scope
	role, _ := openidInfo["role"].(string)
	infos["role"] = role
	sub, _ := claims["sub"].(string)
	infos["svc"] = sub
	scope, _ := openidInfo["scope"].(string)
	infos["scope"] = scope

	// Use role, sub, and scope as needed
	fmt.Println("Role:", role)
	fmt.Println("Sub:", sub)
	fmt.Println("Scope:", scope)

	return infos, nil
}
