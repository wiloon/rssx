package user

import (
	"rssx/common"
	"rssx/utils"
	"rssx/utils/logger"
	log "rssx/utils/logger"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// LegacyUserId is the placeholder owner of all subscriptions and read state from
// before RSSX was multi-user. Only the startup migration (package legacy) may
// use it; request handlers take the user id from the JWT.
const LegacyUserId = "0"

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

// Register stores a new user with a fresh UUID. The name column is unique, so a
// concurrent registration of the same name fails here.
func (u *User) Register() error {
	u.CreateTime = utils.CurrentDateString()
	u.Id = uuid.New().String()
	return common.DB.Create(u).Error
}

func (u *User) Validate() bool {
	pass := false
	tmp := &User{}
	common.DB.Where("name = ?", u.Name).First(tmp)
	log.Debugf("user from db, params: %+v", tmp)

	if tmp.Password != "" {
		err := bcrypt.CompareHashAndPassword([]byte(tmp.Password), []byte(u.Password))
		if err == nil {
			pass = true
			u.Id = tmp.Id
		}
	}
	return pass
}
