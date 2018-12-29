package handle

import (
	"context"
	"github.com/go-chi/chi"
	"github.com/jinzhu/gorm"
	"github.com/ztc1997/teaching_info_system_backend/model"
	"log"
	"net/http"
	"strconv"
	"time"
)

type ProjectForm struct {
	Id         uint    `json:"id"`
	Name       string  `json:"name"`
	Principal  string  `json:"principal"`
	FundsTotal float64 `json:"fundsTotal"`
	FundsUsed  float64 `json:"fundsUsed"`
	Deadline   string  `json:"deadline"`
}

type ProjectResult ProjectForm

func ProjectCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var form ProjectForm
		err := json.NewDecoder(r.Body).Decode(&form)
		r.Body.Close()
		if err != nil || len(form.Name) == 0 || len(form.Principal) == 0 || len(form.Deadline) == 0 {
			log.Printf("fail to parse form: %v", err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), "form", form)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func CreateProject(w http.ResponseWriter, r *http.Request) {
	form := r.Context().Value("form").(ProjectForm)

	project := model.Project{}
	err := form.toProjectModel(&project)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	project.User = r.Context().Value("user").(model.User)

	err = project.Create()
	if err != nil {
		log.Printf("fail to create project: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	projectResult := ProjectResult{}
	projectResult.ParseProjectModel(project)

	writeResult(w, projectResult)
}

func DeleteProject(w http.ResponseWriter, r *http.Request) {
	projectId, err := parseProjectId(r)
	if err != nil {
		log.Printf("fail to parseProjectId: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var project model.Project
	err = project.GetById(projectId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			writeErrorResult(w, "找不到课题")
			return
		}
		log.Printf("fail to get project: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// 权限检查，用户只能删除属于自己的课题
	user := r.Context().Value("user").(model.User)
	if project.UserId != user.ID {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	project.Delete()

	writeResult(w)
}

func UndoDeleteProject(w http.ResponseWriter, r *http.Request) {
	projectId, err := parseProjectId(r)
	if err != nil {
		log.Printf("fail to parseProjectId: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var project model.Project
	err = project.UnscopedGetById(projectId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			writeErrorResult(w, "找不到课题")
			return
		}
		log.Printf("fail to get project: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if project.DeletedAt == nil {
		writeErrorResult(w, "课题未被删除")
		return
	}

	// 权限检查
	user := r.Context().Value("user").(model.User)
	if project.UserId != user.ID {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	project.UndoDelete()

	result := ProjectResult{}
	result.ParseProjectModel(project)
	writeResult(w, result)
}

func SaveProject(w http.ResponseWriter, r *http.Request) {
	form := r.Context().Value("form").(ProjectForm)

	var project model.Project
	err := project.GetById(form.Id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			writeErrorResult(w, "找不到课题")
			return
		}
		log.Printf("fail to get project: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// 权限检查，用户只更新除属于自己的课题
	user := r.Context().Value("user").(model.User)
	if project.UserId != user.ID {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	form.toProjectModel(&project)

	err = project.Save()
	if err != nil {
		log.Printf("fail to update project: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	result := ProjectResult{}
	result.ParseProjectModel(project)
	writeResult(w, result)
}

func GetProjects(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(model.User)
	projects, err := model.DefaultProject.GetProjects(user.ID)
	if err != nil {
		log.Printf("fail to get projects: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	projectResults := make([]ProjectResult, len(projects))
	for i, p := range projects {
		projectResults[i].ParseProjectModel(p)
	}

	writeResult(w, projectResults)
}

func (f *ProjectForm) toProjectModel(m *model.Project) (err error) {
	m.Name = f.Name
	m.Principal = f.Principal
	m.FundsTotal = f.FundsTotal
	m.FundsUsed = f.FundsUsed
	m.Deadline, err = time.Parse("2006-01-02", f.Deadline)
	return
}

func (r *ProjectResult) ParseProjectModel(m model.Project) {
	if m.Model != nil {
		r.Id = m.ID
	}
	r.Name = m.Name
	r.Principal = m.Principal
	r.FundsTotal = m.FundsTotal
	r.FundsUsed = m.FundsUsed
	r.Deadline = m.Deadline.Format("2006-01-02")
	return
}

func parseProjectId(r *http.Request) (projectId uint, err error) {
	projectIdStr := chi.URLParam(r, "projectId")
	i, err := strconv.ParseUint(projectIdStr, 10, strconv.IntSize)
	if err != nil {
		return
	}

	projectId = uint(i)
	return
}
