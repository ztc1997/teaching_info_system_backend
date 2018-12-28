package handle

import (
	"github.com/go-chi/chi"
	"github.com/json-iterator/go"
	"github.com/ztc1997/teaching_info_system_backend/model"
	"log"
	"net/http"
	"strconv"
)

type UserForm struct {
	Username string `json:"username"`
	Password string `json:"password"`
	UserType int    `json:"userType"`
}

type SetPasswordForm struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

type UserResult struct {
	Id       uint   `json:"id"`
	Username string `json:"username"`
	UserType int    `json:"userType"`
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(model.User)
	result := UserResult{}
	result.ParseUserModel(user)
	writeResult(w, result)
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	var userId uint
	{
		projectIdStr := chi.URLParam(r, "userId")
		i, err := strconv.ParseUint(projectIdStr, 10, strconv.IntSize)
		if err != nil {
			log.Printf("fail to ParseUint: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		userId = uint(i)
	}

	err := model.DefaultUser.DeleteById(userId)
	if err != nil {
		log.Printf("fail to delete user: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	writeResult(w)

	err = model.DefaultLoginToken.LogoutAll(userId)
	if err != nil {
		log.Printf("fail to LogoutAll: %v", err)
	}

	err = model.DefaultProject.DeleteByUserId(userId)
	if err != nil {
		log.Printf("fail to delete project by user id: %v", err)
	}
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	var form UserForm
	err := jsoniter.ConfigFastest.NewDecoder(r.Body).Decode(&form)
	r.Body.Close()
	if err != nil || len(form.Username) == 0 || len(form.Username) > 20 || len(form.Password) < 8 {
		log.Printf("fail to parse form: %v", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	user := model.User{Username: form.Username, UserType: form.UserType}

	err = user.SetPassword(form.Password)
	if err != nil {
		log.Printf("fail to set user password： %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = user.Create()
	if err != nil {
		if err == model.ErrUserNameExisted {
			writeErrorResult(w, "用户名已存在")
			return
		}
		log.Printf("fail to CreateUser： %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	result := UserResult{}
	result.ParseUserModel(user)
	writeResult(w, result)
}

func SetPassword(w http.ResponseWriter, r *http.Request) {
	var form SetPasswordForm
	err := jsoniter.ConfigFastest.NewDecoder(r.Body).Decode(&form)
	r.Body.Close()
	if err != nil || len(form.CurrentPassword) < 8 || len(form.NewPassword) < 8 {
		log.Printf("fail to parse form: %v", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	user := r.Context().Value("user").(model.User)

	err = user.ComparePassword(form.CurrentPassword)
	if err != nil {
		writeErrorResult(w, "当前密码错误")
		return
	}

	err = user.SetPassword(form.NewPassword)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = user.Save()

	writeResult(w)

	err = model.DefaultLoginToken.LogoutAll(user.ID)
	if err != nil {
		log.Printf("fail to LogoutAll: %v", err)
		return
	}
}

func GetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := model.DefaultUser.GetUsers()
	if err != nil {
		log.Printf("fail to get users: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	userResults := make([]UserResult, len(users))
	for i, p := range users {
		userResults[i].ParseUserModel(p)
	}

	writeResult(w, userResults)
}

func (r *UserResult) ParseUserModel(m model.User) {
	if m.Model != nil {
		r.Id = m.ID
	}
	r.Username = m.Username
	r.UserType = m.UserType
	return
}
