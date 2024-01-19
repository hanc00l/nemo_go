package db

import (
	"time"
)

type WikiDocs struct {
	Id             int       `gorm:"primaryKey"`
	SpaceID        string    `gorm:"column:space_id"`
	NodeToken      string    `gorm:"column:node_token"`
	ObjType        string    `gorm:"column:obj_type"`
	ObjToken       string    `gorm:"column:obj_token"`
	Title          string    `gorm:"column:title"`
	Comment        string    `gorm:"column:comment"`
	PinIndex       int       `gorm:"column:pin_index"`
	CreateDatetime time.Time `gorm:"column:create_datetime"`
	UpdateDatetime time.Time `gorm:"column:update_datetime"`
}

func (*WikiDocs) TableName() string {
	return "wiki_docs"
}

// Get 根据ID查询记录
func (w *WikiDocs) Get() (success bool) {
	db := GetDB()
	defer CloseDB(db)

	if result := db.First(w, w.Id); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// GetByNodeToken 根据node_token查询记录
func (w *WikiDocs) GetByNodeToken() (success bool) {
	db := GetDB()
	defer CloseDB(db)

	if result := db.Where("node_token", w.NodeToken).First(w); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// Gets 根据指定的条件，查询满足要求的记录
func (w *WikiDocs) Gets(searchMap map[string]interface{}, page, rowsPerPage int) (results []WikiDocs, count int) {
	orderBy := "pin_index desc,create_datetime"

	db := GetDB()
	defer CloseDB(db)
	for column, value := range searchMap {
		db = db.Where(column, value)
	}
	db = db.Model(w)
	//统计满足条件的总记录数
	var total int64
	db.Count(&total)
	//获取分页查询结果
	if rowsPerPage > 0 && page > 0 {
		db = db.Offset((page - 1) * rowsPerPage).Limit(rowsPerPage)
	}
	db.Order(orderBy).Find(&results)
	return results, int(total)
}

// GetsByIpOrDomain 根据指定的IP或Domain条件，查询满足要求的记录
func (w *WikiDocs) GetsByIpOrDomain(ipId, domainId int) (results []WikiDocs) {
	orderBy := "pin_index desc,create_datetime"
	db := GetDB()
	defer CloseDB(db)

	if ipId > 0 {
		dbDocIp := GetDB().Model(&WikiDocsIP{}).Select("doc_id").Where("ip_id = ?", ipId)
		db = db.Where("id in (?)", dbDocIp)
		CloseDB(dbDocIp)
	}
	if domainId > 0 {
		dbDocDomain := GetDB().Model(&WikiDocsDomain{}).Select("doc_id").Where("domain_id = ?", domainId)
		db = db.Where("id in (?)", dbDocDomain)
		CloseDB(dbDocDomain)
	}

	db = db.Model(w)
	db.Order(orderBy).Find(&results)

	return
}

// Count 统计指定查询条件的记录数量
func (w *WikiDocs) Count(searchMap map[string]interface{}) (count int) {
	db := GetDB()
	defer CloseDB(db)

	for column, value := range searchMap {
		db = db.Where(column, value)
	}
	var result int64
	db.Model(w).Count(&result)

	return int(result)
}

// Add 插入一条新的记录，返回主键ID及成功标志
func (w *WikiDocs) Add() (success bool) {
	//由于wiki的编辑是调用feishu的接口，所以这里不更新create_datetime和update_datetime
	//w.CreateDatetime = time.Now()
	//w.UpdateDatetime = time.Now()

	db := GetDB()
	defer CloseDB(db)
	if result := db.Create(w); result.RowsAffected == 1 {
		return true
	} else {
		return false
	}
}

// Update 更新指定ID的一条记录，列名和内容位于map中
func (w *WikiDocs) Update(updatedMap map[string]interface{}) (success bool) {
	// 由于wiki的编辑是调用feishu的接口，所以这里不更新update_datetime
	//updatedMap["update_datetime"] = time.Now()

	db := GetDB()
	defer CloseDB(db)
	if result := db.Model(w).Updates(updatedMap); result.RowsAffected == 1 {
		return true
	} else {
		return false
	}
}

// Delete 删除指定主键ID的一条记录
func (w *WikiDocs) Delete() (success bool) {
	db := GetDB()
	defer CloseDB(db)
	if result := db.Delete(w, w.Id); result.RowsAffected == 1 {
		return true
	} else {
		return false
	}
}
