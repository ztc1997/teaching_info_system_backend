package main

import (
	"flag"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/ztc1997/teaching_info_system_backend/handle"
	mymiddleware "github.com/ztc1997/teaching_info_system_backend/middleware"
	"github.com/ztc1997/teaching_info_system_backend/model"
	"gopkg.in/ini.v1"
	"io"
	"log"
	"net/http"
	"os"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	configFlag := flag.String("config", "conf.ini", "配置文件的路径")
	flag.Parse()

	cfg, err := ini.Load(*configFlag)
	if err != nil {
		log.Fatalf("Fail to read file: %v\n", err)
	}

	logFilePath := cfg.Section("log").Key("error").String()
	if len(logFilePath) > 0 {
		logFile, err := os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
		defer logFile.Close()

		log.SetOutput(io.MultiWriter(os.Stdout, logFile))
	}

	err = initDb(cfg)
	if err != nil {
		log.Fatalf("fail to InitDB: %v\n", err)
	}
	defer model.CloseDB()

	serveHTTP(cfg)
}

func initDb(cfg *ini.File) (err error) {
	dialect := cfg.Section("database").Key("dialect").String()
	uri := cfg.Section("database").Key("uri").String()
	err = model.InitDB(dialect, uri)
	return
}

func serveHTTP(cfg *ini.File) {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(mymiddleware.Authorize)

	r.Route("/api", func(r chi.Router) {
		r.Post("/login", handle.Login)
		r.Post("/logout", handle.Logout)

		r.Route("/user", func(r chi.Router) {
			r.Get("/", handle.GetUser)
			r.With(mymiddleware.AdminOnly).Post("/", handle.CreateUser)
			r.With(mymiddleware.AdminOnly).Delete("/{userId}", handle.DeleteUser)
		})

		r.With(mymiddleware.AdminOnly).Get("/users", handle.GetUsers)

		r.Post("/setPassword", handle.SetPassword)

		r.Route("/project", func(r chi.Router) {
			r.With(handle.ProjectCtx).Post("/", handle.CreateProject)
			r.With(handle.ProjectCtx).Put("/", handle.SaveProject)
			r.Delete("/{projectId}", handle.DeleteProject)
			r.Post("/undoDelete/{projectId}", handle.UndoDeleteProject)
		})

		r.Get("/projects", handle.GetProjects)
	})

	listen := cfg.Section("server").Key("listen").MustString("127.0.0.1:9000")
	fmt.Printf("listen on: %s\n", listen)
	err := http.ListenAndServe(listen, r)
	if err != nil {
		log.Printf("serve http failed:%v\n", err)
	}
}
