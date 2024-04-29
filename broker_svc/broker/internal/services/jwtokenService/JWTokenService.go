package jwtokenService

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt"
	"gitlab.com/grpasr/asonrythme/broker_svc/broker/internal/config"
	e "gitlab.com/grpasr/common/errors/json"
	obs "gitlab.com/grpasr/common/observability"
	"strings"
)

// JWTAccessGenerate generate the jwt access token
type JWTokenService struct {
	signedKey []byte // secret key
}

func NewJWTokenService(conf *config.Config) *JWTokenService {
	signedKey := conf.TOKENGetSignedKey()
	return &JWTokenService{
		signedKey: []byte(signedKey),
	}
}

// JWTokenIsValidToken valid the jwt_token(only for APIserver), if valid return nil
func (a *JWTokenService) JWTokenIsValidToken(ctx context.Context, tokenString string) (map[string]string, e.IError) {

	var secretKey = a.signedKey
	tokenString = strings.TrimSpace(tokenString)
	tokenString = strings.Trim(tokenString, `"`)

	var claims jwt.MapClaims
	var ok bool

	_, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {

		claims, ok = token.Claims.(jwt.MapClaims)
		if !ok {
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
			obs.Logging.NewLogHandler(obs.Logging.LLHError()).
				Err(err).
				Msg("json token expired")
			return nil, e.NewCustomHTTPStatus(e.StatusUnauthorized)
		}
		return nil, e.NewCustomHTTPStatus(e.StatusForbidden)
	}

	// Token is valid process it
	// Extracting the nested "openidInfo" map
	openidInfo, ok := claims["openidInfo"].(map[string]interface{})
	if !ok {
		obs.Logging.NewLogHandler(obs.Logging.LLHError()).
			Err(err).
			Msg("extracting openidInfo from token failed")
		return nil, e.NewCustomHTTPStatus(e.StatusForbidden)
	}

	infos := make(map[string]string)

	// Extracting role, sub, and scope
	var errData = false
	role, ok := openidInfo["role"].(string)
	infos["role"] = role
	if !ok {
		obs.Logging.NewLogHandler(obs.Logging.LLHError()).
			Err(err).
			Msg("no 'role' field in openidInfo")
		errData = true
	}

	sub, ok := claims["sub"].(string)
	infos["svc"] = sub
	if !ok {
		obs.Logging.NewLogHandler(obs.Logging.LLHError()).
			Err(err).
			Msg("no 'sub' field in openidInfo")
		errData = true
	}

	scope, ok := openidInfo["scope"].(string)
	infos["scope"] = scope
	if !ok {
		obs.Logging.NewLogHandler(obs.Logging.LLHError()).
			Err(err).
			Msg("no 'scope' field in openidInfo")
		errData = true
	}

	if errData {
		return nil, e.NewCustomHTTPStatus(e.StatusForbidden, "", "token contains invalid data")
	}

	return infos, nil
}
