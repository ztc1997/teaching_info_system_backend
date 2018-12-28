package middleware

import (
	"context"
	"github.com/ztc1997/teaching_info_system_backend/model"
	"log"
	"net/http"
)

func Authorize(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// 忽略登录 API，否则永远无法登录
		if r.URL.Path == "/api/login" {
			next.ServeHTTP(w, r)
			return
		}

		tokenCookie, err := r.Cookie("login-token")
		if err != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		CSRFTokenStr := r.Header.Get("X-CSRF-TOKEN")

		token := model.LoginToken{}

		err = token.ParseTokensStr(tokenCookie.Value, CSRFTokenStr)

		err = token.Check()
		if err != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		err = token.GetUser()
		if err != nil {
			log.Printf("fail to GetUser: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(r.Context(), "user", token.User)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
