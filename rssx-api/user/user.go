package user

import (
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"rssx/common"
	"rssx/utils"
	"rssx/utils/logger"
)

const DefaultId = "0"

type User struct {
	Id         string
	Name       string
	Password   string
	CreateTime string
}

func (u *User) getByName() {
	common.DB.Where("name = ?", u.Name).First(u)
	logger.Debugf("is exist, user: %v", u)
}
func (u *User) IsExist() bool {
	exist := false
	tmp := &User{}
	common.DB.Where("name = ?", u.Name).First(tmp)
	logger.Debugf("is exist, user: %v", tmp)
	if tmp.Id != "" {
		exist = true
	}
	return exist
}

func (u *User) Register() {
	u.CreateTime = utils.CurrentDateString()
	u.Id = uuid.NewV4().String()
	common.DB.Create(u)
}

func (u *User) Validate() bool {
	pass := false
	tmp := &User{}
	common.DB.Where("name = ?", u.Name).First(tmp)

	if tmp.Password != "" {
		err := bcrypt.CompareHashAndPassword([]byte(tmp.Password), []byte(u.Password))
		if err == nil {
			pass = true
		}
	}
	return pass
}
