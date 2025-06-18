package xjwt

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

type TokenConf struct {
	AccessSecret  string
	AccessExpire  int64
	RefreshSecret string
	RefreshExpire int64
	RefreshAfter  int64
}

var ErrGenerateTokenError = errors.New("generate token error")

type (
	TokenOptions struct {
		AccessSecret string
		AccessExpire int64
		RefreshAfter int64
		Fields       map[string]interface{}
	}

	Token struct {
		AccessToken  string `json:"access_token"`
		AccessExpire int64  `json:"access_expire"`
		RefreshAfter int64  `json:"refresh_after"`
	}

	CustomClaims struct {
		jwt.StandardClaims
		Fields map[string]interface{}
	}
)

// BuildTokens generates and returns an access token with an expiration time.
func BuildTokens(opt TokenOptions) (Token, error) {
	var token Token
	now := time.Now().Unix()
	accessToken, err := genToken(now, opt.AccessSecret, opt.Fields, opt.AccessExpire)
	if err != nil {
		return token, err
	}
	token.AccessToken = accessToken
	token.AccessExpire = now + opt.AccessExpire
	token.RefreshAfter = now + opt.AccessExpire/2
	if opt.RefreshAfter > 0 {
		token.RefreshAfter = now + opt.RefreshAfter
	}

	return token, nil
}

// genToken generates a signed JWT token with custom claims.
func genToken(iat int64, secretKey string, payloads map[string]interface{}, seconds int64) (string, error) {
	claims := make(jwt.MapClaims)
	// Token expiration time
	claims["exp"] = iat + seconds
	// Token issued at time
	claims["iat"] = iat
	// Custom payloads
	for k, v := range payloads {
		claims[k] = v
	}

	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = claims

	return token.SignedString([]byte(secretKey))
}

// ValidateToken parses and validates a JWT token.
func ValidateToken(tokenString, secretKey string) (map[string]interface{}, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Check if the signing method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	// If token is valid, extract claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// ExtractClaims extracts custom claims from a token string.
func ExtractClaims(tokenString, secretKey string) (map[string]interface{}, error) {
	claims, err := ValidateToken(tokenString, secretKey)
	if err != nil {
		return nil, err
	}

	// Extract custom fields from claims
	result := make(map[string]interface{})
	for k, v := range claims {
		// Skip standard claims (exp, iat, etc.)
		if k == "exp" || k == "iat" {
			continue
		}
		result[k] = v
	}

	return result, nil
}
