package middleware

import (
	"github.com/ztc1997/teaching_info_system_backend/model"
	"net/http"
)

func AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value("user").(model.User)

		if user.UserType != model.UserTypeAdmin {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
