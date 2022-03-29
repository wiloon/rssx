package jwt

import (
	"fmt"
	"github.com/golang-jwt/jwt"
	uuid "github.com/satori/go.uuid"
	"rssx/utils"
	"testing"
	"time"
)

func TestGenJwtToken(t *testing.T) {
	id := "foo"
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
	signedString, err := token.SignedString([]byte("fa6ee430-ebf6-44f2-a18b-64d691cd2dae"))
	fmt.Printf("token: %v, err: %v", signedString, err)
}
