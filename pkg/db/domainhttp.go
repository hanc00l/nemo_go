package db

import (
	"time"
)

const HttpBodyContentSize = 16000 // 由于utf-8-mb4的mysql每行sizew不能超过16383（64K/4）

type DomainHttp struct {
	Id             int       `gorm:"primaryKey"`
	RelatedId      int       `gorm:"column:r_id"`
	Port           int       `gorm:"column:port"`
	Source         string    `gorm:"column:source"`
	Tag            string    `gorm:"column:tag"`
	Content        string    `gorm:"column:content"`
	CreateDatetime time.Time `gorm:"column:create_datetime"`
	UpdateDatetime time.Time `gorm:"column:update_datetime"`
}

func (DomainHttp) TableName() string {
	return "domain_http"
}

// Add 插入一条新的记录，返回主键ID及成功标志
func (d *DomainHttp) Add() (success bool) {
	d.CreateDatetime = time.Now()
	d.UpdateDatetime = time.Now()

	db := GetDB()
	defer CloseDB(db)
	if result := db.Create(d); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// GetsByRelatedId 根据查询条件执行数据库查询操作，返回查询结果数组
func (d *DomainHttp) GetsByRelatedId() (results []DomainHttp) {
	orderBy := "tag,update_datetime desc"

	db := GetDB()
	defer CloseDB(db)
	db.Where("r_id", d.RelatedId).Order(orderBy).Find(&results)
	return
}

// GetsByRelatedIdAndTag 根据查询条件执行数据库查询操作，返回查询结果数组
func (d *DomainHttp) GetsByRelatedIdAndTag() (results []DomainHttp) {
	orderBy := "tag,update_datetime desc"

	db := GetDB()
	defer CloseDB(db)
	db.Where("r_id", d.RelatedId).Where("tag", d.Tag).Order(orderBy).Find(&results)
	return
}

// GetByRelatedIdAndPortAndTag 根据端口、port、tag查询一条记录
func (d *DomainHttp) GetByRelatedIdAndPortAndTag() (success bool) {
	db := GetDB()
	defer CloseDB(db)

	if result := db.Where("r_id", d.RelatedId).Where("port", d.Port).Where("tag", d.Tag).First(d); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// Update 更新指定ID的一条记录，列名和内容位于map中
func (d *DomainHttp) Update(updateMap map[string]interface{}) (success bool) {
	updateMap["update_datetime"] = time.Now()

	db := GetDB()
	defer CloseDB(db)
	if result := db.Model(d).Updates(updateMap); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// Delete 删除指定主键ID的一条记录
func (d *DomainHttp) Delete() (success bool) {
	db := GetDB()
	defer CloseDB(db)
	if result := db.Delete(d, d.Id); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// SaveOrUpdate 保存、更新一条记录
func (d *DomainHttp) SaveOrUpdate() (success bool) {
	if d.GetByRelatedIdAndPortAndTag() {
		updateMap := make(map[string]interface{})
		updateMap["source"] = d.Source
		updateMap["tag"] = d.Tag
		updateMap["content"] = d.Content
		return d.Update(updateMap)
	} else {
		return d.Add()
	}
}
