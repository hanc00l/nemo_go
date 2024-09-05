package db

import (
	"time"
)

type WikiDocsIP struct {
	Id             int       `gorm:"primaryKey"`
	DocumentId     int       `gorm:"column:doc_id"`
	IpId           int       `gorm:"column:ip_id"`
	CreateDatetime time.Time `gorm:"column:create_datetime"`
	UpdateDatetime time.Time `gorm:"column:update_datetime"`
}

func (*WikiDocsIP) TableName() string {
	return "wiki_docs_ip"
}

func (w *WikiDocsIP) Get() (success bool) {
	db := GetDB()
	defer CloseDB(db)

	if result := db.First(w, w.Id); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

func (w *WikiDocsIP) RemoveByDocId(documentId int) (success bool) {
	db := GetDB()
	defer CloseDB(db)

	if result := db.Where("doc_id", documentId).Delete(w); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

func (w *WikiDocsIP) Add() (success bool) {
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
