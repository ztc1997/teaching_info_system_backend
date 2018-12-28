package model

import (
	"bytes"
	"encoding/base64"
	"errors"
	"github.com/jinzhu/gorm"
	"github.com/satori/go.uuid"
	"log"
	"time"
)

// 登录有效期180天
const LoginPeriod = 180 * 24 * time.Hour

var ErrCSRFTokenMismatch = errors.New("CSRFToken 不匹配")

type LoginToken struct {
	Token []byte `gorm:"primary_key;type:BINARY(16)"`

	// CSRFToken 由前端保存，并附加到请求，确保请求发自正确的网站
	CSRFToken []byte `gorm:"type:BINARY(16)"`

	CreatedAt time.Time
	Expires   time.Time
	UserId    uint
	User      User
}

var DefaultLoginToken = new(LoginToken)

// 检查 token 是否有效
func (t *LoginToken) Check() (err error) {
	CSRFToken := t.CSRFToken
	err = db.First(t).Error
	if err != nil {
		return
	}
	if !bytes.Equal(t.CSRFToken, CSRFToken) {
		return ErrCSRFTokenMismatch
	}
	return
}

func (t *LoginToken) GetUser() (err error) {
	err = db.Model(t).Related(&t.User).Error
	return
}

func (t *LoginToken) Login() (err error) {
	t.Token = uuid.Must(uuid.NewV4()).Bytes()
	t.CSRFToken = uuid.Must(uuid.NewV4()).Bytes()
	t.Expires = gorm.NowFunc().Add(LoginPeriod)
	err = db.Create(t).Error
	return
}

func (t *LoginToken) Logout() (err error) {
	err = db.Delete(t).Error
	return
}

func (t *LoginToken) LogoutAll(userId uint) (err error) {
	err = db.Delete(t, "user_id = ?", userId).Error
	return
}

func (t *LoginToken) GetTokensStr() (token, CSRFToken string) {
	token = base64.StdEncoding.EncodeToString(t.Token)[:22]
	CSRFToken = base64.StdEncoding.EncodeToString(t.CSRFToken)[:22]
	return
}

func (t *LoginToken) ParseTokensStr(token, CSRFToken string) (err error) {
	t.Token, err = base64.StdEncoding.DecodeString(token + "==")
	if err != nil {
		return
	}
	t.CSRFToken, err = base64.StdEncoding.DecodeString(CSRFToken + "==")
	return
}

func (t *LoginToken) ClearUp() (err error) {
	err = db.Delete(t, "expires < ?", time.Now()).Error
	return
}

func (t *LoginToken) ScheduleClearUp() (quit chan struct{}) {
	quit = make(chan struct{})
	ticker := time.NewTicker(time.Minute * 1)
	go func() {
		for {
			select {
			case <-ticker.C:
				err := t.ClearUp()
				if err != nil {
					log.Printf("fail to ClearUp: %v", err)
				}
			case <-quit:
				return
			}
		}
	}()
	return
}
