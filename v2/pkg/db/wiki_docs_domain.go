package db

import (
	"time"
)

type WikiDocsDomain struct {
	Id             int       `gorm:"primaryKey"`
	DocumentId     int       `gorm:"column:doc_id"`
	DomainId       int       `gorm:"column:domain_id"`
	CreateDatetime time.Time `gorm:"column:create_datetime"`
	UpdateDatetime time.Time `gorm:"column:update_datetime"`
}

func (*WikiDocsDomain) TableName() string {
	return "wiki_docs_domain"
}

func (w *WikiDocsDomain) Get() (success bool) {
	db := GetDB()
	defer CloseDB(db)

	if result := db.First(w, w.Id); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

func (w *WikiDocsDomain) RemoveByDocument(documentId int) (success bool) {
	db := GetDB()
	defer CloseDB(db)

	if result := db.Where("doc_id", documentId).Delete(w); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

func (w *WikiDocsDomain) Add() (success bool) {
	w.CreateDatetime = time.Now()
	w.UpdateDatetime = time.Now()

	db := GetDB()
	defer CloseDB(db)

	if result := db.Create(w); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}
