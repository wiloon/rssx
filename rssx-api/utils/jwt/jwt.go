package jwt

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/patrickmn/go-cache"
	"github.com/satori/go.uuid"
	"rssx/utils"
	"rssx/utils/config"
	"rssx/utils/logger"
	"rssx/utils/response"
	"strings"
	"time"
)

var tokenRefreshCache *cache.Cache

func init() {
	tokenRefreshCache = cache.New(1*time.Minute, 10*time.Minute)
}

type RssxClaims struct {
	Id string `json:"id"`
	jwt.StandardClaims
}

func NewToken(id string) string {
	// Create the Claims
	claims := RssxClaims{
		id,
		jwt.StandardClaims{
			Audience:  "rssx.wiloon.net",
			ExpiresAt: utils.DateToSeconds(time.Now().AddDate(0, 0, 1)),
			Id:        uuid.NewV4().String(),
			IssuedAt:  utils.CurrentSeconds(),
			Issuer:    "wiloon.com",
			NotBefore: utils.CurrentSeconds(),
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

func GetJwtToken(jwtPayload Payload) (token string, err error) {
	jwtToken := jwt.NewWithClaims(
		jwt.SigningMethodHS512, // method
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

const DefaultIss = "wiloon.com"
const DefaultSub = "rssx"

type Payload struct {
	Iss string // (issuer)：签发人
	Sub string // (subject)：主题
	Aud string // (audience)：受众
	Nbf int64  // (Not Before)：生效时间
	Exp int64  // (expiration time)：过期时间
	Iat int64  // (Issued At)：签发时间
	Jti string // (JWT ID)：编号
	Id  string // user id

}

// ParseToken signature is invalid
// Token is expired
func ParseToken(tokenString string) (jwtPayload *Payload, err error) {
	// Parse takes the token string and a function for looking up the key. The latter is especially
	// useful if you use multiple keys for your application.  The standard is to use 'kid' in the
	// head of the token to identify which key to use, but the parsed token (head and claims) is provided
	// to the callback, providing flexibility.
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return config.GetString("security-key", ""), nil
	})

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok {
		err = errors.New("cannot convert claim to mapclaim")
		return
	}
	//验证token，如果token被修改过则为false
	if !token.Valid {
		err = errors.New("token is invalid")
		return
	}
	if claims["iss"] != nil {
		jwtPayload.Iss = claims["iss"].(string)
	}
	if claims["sub"] != nil {
		jwtPayload.Sub = claims["sub"].(string)
	}
	if claims["aud"] != nil {
		jwtPayload.Sub = claims["aud"].(string)
	}
	if claims["nbf"] != nil {
		jwtPayload.Nbf = int64(claims["nbf"].(float64))
	}
	if claims["exp"] != nil {
		jwtPayload.Exp = int64(claims["exp"].(float64))
	}
	if claims["iat"] != nil {
		jwtPayload.Iat = int64(claims["iat"].(float64))
	}
	if claims["jti"] != nil {
		jwtPayload.Jti = claims["jti"].(string)
	}
	if claims["id"] != nil {
		jwtPayload.Id = claims["id"].(string)
	}
	return jwtPayload, err
}

func secret() jwt.Keyfunc {
	return func(token *jwt.Token) (i interface{}, e error) {
		keyBytes, _ := base64.RawURLEncoding.DecodeString(config.GetString("security-key", ""))
		return []byte(keyBytes), nil
	}
}

// token刷新, 处理token快过期的时候 自动 刷新 ,防止用户操作中断
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

func GetJwtTokenFromHeader(c *gin.Context) string {
	token := ""
	tokenStr := c.GetHeader("Authorization")
	tokenStr = strings.TrimSpace(tokenStr)
	arr := strings.Split(tokenStr, "Bearer ")
	if len(arr) >= 2 {
		token = arr[1]
	}
	// logger.Debugf("get token from header, url: %s, token: %s", c.Request.RequestURI, token)
	return token
}

func GetUserId(c *gin.Context) string {
	token := GetJwtTokenFromHeader(c)
	p, err := ParseToken(token)
	if err != nil {
		logger.Error("failed to parse token: %v", err)
	}

	return p.Id
}

func GetId(c *gin.Context) string {
	p := GetJwtPayLoad(c)
	if p != nil {
		return p.Id
	}
	return ""
}

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

func TokenNotExist(c *gin.Context) bool {
	return GetJwtTokenFromHeader(c) == ""
}

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
		// refresh token
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
		// refresh token
		logger.Debugf("refresh token, user type: %v, uuid: %v, open id: %v", p.Id)
		newToken = New(p.Id)
		tokenRefreshCache.Set(redisKey, time.Now(), cache.DefaultExpiration)
		logger.Infof("token refreshed, duration till exp: %v,new token: %s", d0, newToken)
	} else {
		logger.Debugf("valid token, refresh ignore, duration till exp: %v", d0)
	}
	return newToken, nil
}

func New(id string) string {
	tokenDuration, _ := time.ParseDuration("8h")
	jwtPayload := Payload{
		Iss: "wiloon.com",
		Sub: "rssx",
		Nbf: utils.CurrentSeconds(),
		Exp: utils.DateToSeconds(time.Now().Add(tokenDuration)),
		Iat: utils.CurrentSeconds(),
		Jti: uuid.NewV4().String(),

		Id: id,
	}

	token, err := GetJwtToken(jwtPayload)
	if err != nil {
		logger.Error("failed to sign jwt", err)
	}
	return token
}

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
	//验证token，如果token被修改过则为false
	if !token.Valid {
		err = errors.New("token is invalid")
		valid = false
	}
	logger.Infof("token check result, valid: %v", valid)
	return valid
}

type TokenBuilder struct {
	payload Payload
}

func (t *TokenBuilder) makeCommonFiles(id, groupId, dealerId string) {
	t.payload.Iss = DefaultIss
	t.payload.Sub = DefaultSub
	t.payload.Nbf = utils.CurrentSeconds()
	t.payload.Exp = utils.DateToSeconds(time.Now().AddDate(0, 0, 1))
	t.payload.Iat = utils.CurrentSeconds()
	t.payload.Jti = uuid.NewV4().String()
	t.payload.Id = id

}
