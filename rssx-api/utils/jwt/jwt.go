package jwt

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
	"rssx/utils"
	"rssx/utils/config"
	"rssx/utils/logger"
	"rssx/utils/response"
)

var tokenRefreshCache *cache.Cache

func init() {
	tokenRefreshCache = cache.New(1*time.Minute, 10*time.Minute)
}

// RssxClaims is the custom claims structure for RSSX tokens.
// In jwt/v5, we embed jwt.RegisteredClaims instead of using jwt.StandardClaims.
type RssxClaims struct {
	Id string `json:"id"`
	jwt.RegisteredClaims
}

// NewToken generates a new JWT token for the given user ID.
// The token expires in 1 day and uses HS256 signing method.
func NewToken(id string) string {
	claims := RssxClaims{
		Id: id,
		RegisteredClaims: jwt.RegisteredClaims{
			Audience:  jwt.ClaimStrings{"rssx.wiloon.net"},
			ExpiresAt: jwt.NewNumericDate(time.Now().AddDate(0, 0, 1)),
			ID:        uuid.New().String(),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "wiloon.com",
			NotBefore: jwt.NewNumericDate(time.Now()),
			Subject:   "rssx",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedString, err := token.SignedString([]byte(config.GetString("rssx.security-key", "")))
	if err != nil {
		logger.Infof("failed to gen jwt token: %v", err)
	}

	return signedString
}

// Payload represents the parsed JWT payload for API responses.
type Payload struct {
	Iss string // (issuer): issuer
	Sub string // (subject): subject
	Aud string // (audience): audience
	Nbf int64  // (Not Before): not before timestamp
	Exp int64  // (expiration time): expiration timestamp
	Iat int64  // (Issued At): issued at timestamp
	Jti string // (JWT ID): unique token ID
	Id  string // user id
}

const DefaultIss = "wiloon.com"
const DefaultSub = "rssx"

// GetJwtToken creates a new JWT token with the given payload using HS512.
func GetJwtToken(jwtPayload Payload) (token string, err error) {
	jwtToken := jwt.NewWithClaims(
		jwt.SigningMethodHS512,
		jwt.MapClaims{
			"iss": jwtPayload.Iss,
			"sub": jwtPayload.Sub,
			"aud": jwtPayload.Aud,
			"nbf": jwtPayload.Nbf,
			"exp": jwtPayload.Exp,
			"iat": jwtPayload.Iat,
			"jti": jwtPayload.Jti,
			"id":  jwtPayload.Id,
		})
	keyBytes, _ := base64.RawURLEncoding.DecodeString(config.GetString("security-key", ""))
	return jwtToken.SignedString(keyBytes)
}

// ParseToken parses and validates a JWT token string.
// Returns the parsed Payload or an error if the token is invalid.
// Possible errors: token is expired, signature is invalid, or token is malformed.
func ParseToken(tokenString string) (jwtPayload *Payload, err error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return config.GetString("security-key", ""), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("cannot convert claim to mapclaim")
	}

	if !token.Valid {
		return nil, errors.New("token is invalid")
	}

	jwtPayload = &Payload{}

	if iss, ok := claims["iss"].(string); ok {
		jwtPayload.Iss = iss
	}
	if sub, ok := claims["sub"].(string); ok {
		jwtPayload.Sub = sub
	}
	if aud, ok := claims["aud"].(string); ok {
		jwtPayload.Aud = aud
	}
	if nbf, ok := claims["nbf"].(float64); ok {
		jwtPayload.Nbf = int64(nbf)
	}
	if exp, ok := claims["exp"].(float64); ok {
		jwtPayload.Exp = int64(exp)
	}
	if iat, ok := claims["iat"].(float64); ok {
		jwtPayload.Iat = int64(iat)
	}
	if jti, ok := claims["jti"].(string); ok {
		jwtPayload.Jti = jti
	}
	if id, ok := claims["id"].(string); ok {
		jwtPayload.Id = id
	}

	return jwtPayload, nil
}

func secret() jwt.Keyfunc {
	return func(token *jwt.Token) (i interface{}, e error) {
		keyBytes, _ := base64.RawURLEncoding.DecodeString(config.GetString("security-key", ""))
		return keyBytes, nil
	}
}

// RefreshToken handles token refresh when a token is about to expire.
// It prevents user interruption during active usage.
func RefreshToken(c *gin.Context) {
	logger.Debugf("refresh token")
	data := make(map[string]string)
	oldToken := GetJwtTokenFromHeader(c)
	newToken := ""
	if oldToken != "" {
		p := GetJwtPayLoad(c)
		t, err := refreshTokenByExp(p)
		if err != nil {
			newToken = ""
		} else {
			newToken = t
		}
	}
	data["token"] = newToken
	logger.Infof("refresh token, old token: %s", oldToken)
	logger.Infof("refresh token, new token: %s", newToken)
	response.ShowData(c, data)
}

// GetJwtTokenFromHeader extracts the JWT token from the Authorization header.
// Expected format: "Bearer <token>"
func GetJwtTokenFromHeader(c *gin.Context) string {
	token := ""
	tokenStr := c.GetHeader("Authorization")
	tokenStr = strings.TrimSpace(tokenStr)
	arr := strings.Split(tokenStr, "Bearer ")
	if len(arr) >= 2 {
		token = arr[1]
	}
	return token
}

// GetUserId extracts the user ID from the JWT token in the request header.
func GetUserId(c *gin.Context) string {
	token := GetJwtTokenFromHeader(c)
	p, err := ParseToken(token)
	if err != nil {
		logger.Error("failed to parse token: %v", err)
	}

	return p.Id
}

// GetId is an alias for GetUserId.
func GetId(c *gin.Context) string {
	p := GetJwtPayLoad(c)
	if p != nil {
		return p.Id
	}
	return ""
}

// GetJwtPayLoad parses and returns the JWT payload from the request.
func GetJwtPayLoad(c *gin.Context) *Payload {
	token := GetJwtTokenFromHeader(c)
	if token != "" {
		p, err := ParseToken(token)
		if err == nil {
			return p
		}
	}
	return nil
}

// TokenNotExist returns true if no JWT token is present in the request header.
func TokenNotExist(c *gin.Context) bool {
	return GetJwtTokenFromHeader(c) == ""
}

// CheckAndRefreshToken checks if the token needs refresh (expires within 5 minutes)
// and sets the refresh-token header if needed.
func CheckAndRefreshToken(c *gin.Context) *Payload {
	logger.Debugf("check if token need refresh")
	p := GetJwtPayLoad(c)
	logger.Debugf("parsed token: %+v", p)
	needRefresh, err := checkIfTokenNeedRefresh(p)
	if err != nil {
		logger.Warn("ignore duplicate refresh")
		return p
	}
	if needRefresh {
		c.Writer.Header().Set("refresh-token", "true")
		logger.Infof("token need refresh: %v", "yes")
	}
	return p
}

const redisKeyPrefixTokenCheck = "rssx:token:check:"

func checkIfTokenNeedRefresh(p *Payload) (bool, error) {
	tokenNeedRefresh := false
	exp := utils.SecondsToTime(p.Exp)
	d0 := exp.Sub(time.Now())
	redisKey := redisKeyPrefixTokenCheck + p.Id
	if d0 <= 5*time.Minute {
		lastCheckTime, found := tokenRefreshCache.Get(redisKey)
		if found {
			d := lastCheckTime.(time.Time).Sub(time.Now())
			if d <= 1*time.Minute {
				logger.Warnf("ignore duplicate refresh check, duration: %v", d)
				e := errors.New("ignore duplicate refresh check")
				return tokenNeedRefresh, e
			}
		}
		tokenNeedRefresh = true
		tokenRefreshCache.Set(redisKey, time.Now(), cache.DefaultExpiration)
		logger.Infof("token refresh check, token need refresh, duration till exp: %v", d0)
	} else {
		logger.Debugf("token refresh check, valid token, refresh ignore, duration till exp: %v", d0)
	}
	return tokenNeedRefresh, nil
}

const redisKeyPrefixTokenRefresh = "rssx:token:refresh:"

func refreshTokenByExp(p *Payload) (string, error) {
	newToken := ""
	exp := utils.SecondsToTime(p.Exp)
	d0 := time.Now().Sub(exp)
	redisKey := redisKeyPrefixTokenRefresh + p.Id
	if d0 <= 5*time.Minute {
		lastRefreshTime, found := tokenRefreshCache.Get(redisKey)
		if found {
			d := lastRefreshTime.(time.Time).Sub(time.Now())
			if d <= 1*time.Minute {
				logger.Warnf("ignore duplicate refresh request, duration: %v", d)
				e := errors.New("ignore duplicate refresh request")
				return newToken, e
			}
		}
		logger.Debugf("refresh token, user type: %v, uuid: %v, open id: %v", p.Id)
		newToken = New(p.Id)
		tokenRefreshCache.Set(redisKey, time.Now(), cache.DefaultExpiration)
		logger.Infof("token refreshed, duration till exp: %v,new token: %s", d0, newToken)
	} else {
		logger.Debugf("valid token, refresh ignore, duration till exp: %v", d0)
	}
	return newToken, nil
}

// New generates a new JWT token with 8-hour expiration.
// This is used for the RefreshToken flow.
func New(id string) string {
	tokenDuration, _ := time.ParseDuration("8h")
	jwtPayload := Payload{
		Iss: DefaultIss,
		Sub: DefaultSub,
		Nbf: utils.CurrentSeconds(),
		Exp: utils.DateToSeconds(time.Now().Add(tokenDuration)),
		Iat: utils.CurrentSeconds(),
		Jti: uuid.New().String(),
		Id:  id,
	}

	token, err := GetJwtToken(jwtPayload)
	if err != nil {
		logger.Error("failed to sign jwt", err)
	}
	return token
}

// IsValidToken validates the JWT token in the request header.
// Returns true if the token is valid, false otherwise.
func IsValidToken(c *gin.Context) bool {
	valid := true
	defer func() {
		if err := recover(); err != nil {
			logger.Errorf("invalid token, recover: %v", err)
			valid = false
		}
	}()

	tokenStr := GetJwtTokenFromHeader(c)
	logger.Debugf("token from header: %v", tokenStr)
	token, err := jwt.Parse(tokenStr, secret())
	if err != nil {
		logger.Warnf("invalid token: %v", err)
		err = errors.New("invalid token")
		valid = false
	}
	_, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		err = errors.New("cannot convert claim to mapclaim")
		valid = false
	}
	if !token.Valid {
		err = errors.New("token is invalid")
		valid = false
	}
	logger.Infof("token check result, valid: %v", valid)
	return valid
}

// TokenBuilder helps construct JWT tokens with common fields.
type TokenBuilder struct {
	payload Payload
}

// MakeCommonFiles sets up the common payload fields for a token.
func (t *TokenBuilder) MakeCommonFiles(id string) {
	t.payload.Iss = DefaultIss
	t.payload.Sub = DefaultSub
	t.payload.Nbf = utils.CurrentSeconds()
	t.payload.Exp = utils.DateToSeconds(time.Now().AddDate(0, 0, 1))
	t.payload.Iat = utils.CurrentSeconds()
	t.payload.Jti = uuid.New().String()
	t.payload.Id = id
}
