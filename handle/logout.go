package handle

import (
	"github.com/ztc1997/teaching_info_system_backend/model"
	"net/http"
)

func Logout(w http.ResponseWriter, r *http.Request) {
	tokenCookie, _ := r.Cookie("login-token")
	token := model.LoginToken{}
	token.ParseTokensStr(tokenCookie.Value, "")
	err := token.Logout()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
