package user

import (
	"strings"

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
		response.ShowError(c, "Registration failed: invalid request")
		return
	}
	u.Name = strings.TrimSpace(u.Name)
	if u.Name == "" || u.Password == "" {
		response.ShowError(c, "Username and password are required")
		return
	}
	if u.IsExist() {
		logger.Debugf("register failed, user exist, name: %s", u.Name)
		response.ShowError(c, "Username is already taken")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Errorf("register failed, hash password: %v", err)
		response.ShowError(c, "Registration failed")
		return
	}
	u.Password = string(hash)
	if err := u.Register(); err != nil {
		logger.Errorf("register failed, name: %s, err: %v", u.Name, err)
		response.ShowError(c, "Registration failed")
		return
	}

	var data = make(map[string]interface{}, 0)
	data["token"] = jwt.NewToken(u.Id)
	logger.Infof("user registered, name: %s, id: %s", u.Name, u.Id)
	response.ShowData(c, data)
}
