package user

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"rssx/utils/jwt"
	"rssx/utils/logger"
	"rssx/utils/response"
)

func Register(c *gin.Context) {
	var u User
	err := c.BindJSON(&u)
	if err != nil {
		logger.Debugf("register failed: %v", err)
		response.ShowError(c, "注册失败")
		return
	}
	if u.IsExist() {
		logger.Debugf("register failed, user exist, name: %s", u.Name)
		response.ShowError(c, "注册失败")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println(err)
	}
	u.Password = string(hash)
	u.Register()

	jwtTokenString := jwt.NewToken(u.Id)
	var data = make(map[string]interface{}, 0)

	data["token"] = jwtTokenString
	logger.Infof("user login, jwt token generated, token: %s", jwtTokenString)
	response.ShowData(c, data)
}
