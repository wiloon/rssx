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
		response.ShowError(c, "Invalid username or password")
		return
	}
	log.Debugf("user login, name: %s", u.Name)
	if u.Name == "" || u.Password == "" {
		response.ShowError(c, "Username and password are required")
		return
	}
	if u.Validate() {
		jwtTokenString := jwt.NewToken(u.Id)
		var data = make(map[string]interface{}, 0)

		data["token"] = jwtTokenString
		log.Infof("user login, jwt issued, user id: %s", u.Id)
		response.ShowData(c, data)
	} else {
		response.ShowError(c, "Invalid username or password")
	}
	return
}
