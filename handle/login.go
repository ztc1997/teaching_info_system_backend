package handle

import (
	"github.com/ztc1997/teaching_info_system_backend/model"
	"log"
	"net/http"
)

type LoginResult struct {
	CSRFToken string `json:"CSRFToken"`
}

type LoginForm struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Login(w http.ResponseWriter, r *http.Request) {
	form := LoginForm{}
	err := json.NewDecoder(r.Body).Decode(&form)
	defer r.Body.Close()
	if err != nil {
		log.Printf("fail to parse form: %v", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if len(form.Username) == 0 {
		writeErrorResult(w, "用户名不能为空")
		return
	}
	if len(form.Username) > 20 {
		writeErrorResult(w, "用户名不能超过20个字符")
		return
	}

	if len(form.Password) < 8 {
		writeErrorResult(w, "密码必须大于8个字符")
		return
	}

	user := model.User{Username: form.Username}
	err = user.FirstByUsername()
	if err != nil {
		writeErrorResult(w, "用户名或密码错误")
		return
	}

	err = user.ComparePassword(form.Password)
	if err != nil {
		writeErrorResult(w, "用户名或密码错误")
		return
	}

	token := model.LoginToken{User: user}
	err = token.Login()
	if err != nil {
		log.Printf("fail to Login： %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	tokenStr, CSRFToken := token.GetTokensStr()

	http.SetCookie(w, &http.Cookie{
		Name:    "login-token",
		Value:   tokenStr,
		Expires: token.Expires,
		// 预防跨站脚本攻击
		HttpOnly: true,
	})

	writeResult(w, LoginResult{CSRFToken})
}
