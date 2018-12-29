package model

import (
	"github.com/jinzhu/gorm"
	"time"
)

var DefaultProject = new(Project)

type Project struct {
	*gorm.Model
	Name       string
	Principal  string
	FundsTotal float64
	FundsUsed  float64
	Deadline   time.Time `gorm:"type:DATE"`
	UserId     uint
	User       User
}

func (p *Project) Create() (err error) {
	err = db.Create(p).Error
	return
}

func (p *Project) Save() (err error) {
	err = db.Save(p).Error
	return
}

func (p *Project) Delete() (err error) {
	err = db.Delete(p).Error
	return
}

func (p *Project) DeleteByUserId(userId uint) (err error) {
	err = db.Delete(p, "user_id = ?", userId).Error
	return
}

func (p *Project) GetById(id uint) (err error) {
	err = db.First(p, id).Error
	return
}

func (p *Project) UnscopedGetById(id uint) (err error) {
	err = db.Unscoped().First(p, id).Error
	return
}

func (p *Project) UndoDelete() (err error) {
	err = db.Unscoped().Model(p).Update("deleted_at", nil).Error
	return
}

func (p *Project) GetProjects(userId uint) (projects []Project, err error) {
	err = db.Find(&projects, "user_id = ?", userId).Error
	return
}
