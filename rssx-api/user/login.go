package user

import (
	"github.com/gin-gonic/gin"
	"rssx/utils"
	"rssx/utils/jwt"
	log "rssx/utils/logger"
	"rssx/utils/response"
)

func Login(c *gin.Context) {
	defer func() {
		utils.RecoverAndPrintStackTrace()
	}()
	var u User
	err := c.BindJSON(&u)
	if err != nil {
		log.Debugf("login, failed to parse params err: %v", err)
		response.ShowError(c, "用户名或密码格式错误")
		return
	}
	log.Debugf("user login, params: %v", u)
	if u.Name == "" || u.Password == "" {
		response.ShowError(c, "用户名或密码为空")
		return
	}
	if u.Validate() {
		jwtTokenString := jwt.NewToken(u.Id)
		var data = make(map[string]interface{}, 0)

		data["token"] = jwtTokenString
		log.Infof("user login, jwt token generated, token: %s", jwtTokenString)
		response.ShowData(c, data)
	} else {
		response.ShowError(c, "用户名或密码错误")
	}
	return
}
