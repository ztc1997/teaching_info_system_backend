package model

import "github.com/jinzhu/gorm"

var db *gorm.DB
var quitScheduleClearUp chan struct{}

func InitDB(dialect string, args ...interface{}) (err error) {
	db, err = gorm.Open(dialect, args...)
	if err != nil {
		return
	}

	if dialect == "sqlite3" {
		db.DB().SetMaxOpenConns(1)
	}

	err = db.AutoMigrate(&User{}, &LoginToken{}, &Project{}).Error
	if err != nil {
		return
	}

	count, err := DefaultUser.Count()
	if err != nil {
		return
	}

	if count == 0 {
		err = DefaultUser.CreateDefaultAdminUser()
		if err != nil {
			return
		}
	}

	quitScheduleClearUp = DefaultLoginToken.PeriodicCleanup()
	return
}

func CloseDB() (err error) {
	close(quitScheduleClearUp)
	if db == nil {
		panic("close DB without init")
	}
	return db.Close()
}
