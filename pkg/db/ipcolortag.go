package db

import (
	"time"
)

type IpColorTag struct {
	Id             int       `gorm:"primaryKey"`
	RelatedId      int       `gorm:"column:r_id"`
	Color            string    `gorm:"column:color"`
	CreateDatetime time.Time `gorm:"column:create_datetime"`
	UpdateDatetime time.Time `gorm:"column:update_datetime"`
}

func (IpColorTag)TableName()string{
	return "ip_color_tag"
}

//Add 插入一条新的记录
func (ipct *IpColorTag) Add() (success bool){
	ipct.CreateDatetime = time.Now()
	ipct.UpdateDatetime = time.Now()

	db := GetDB()
	defer CloseDB(db)
	if result := db.Create(ipct);result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// GetByRelatedId 根据指定r_id查询一条记录
func (ipct *IpColorTag) GetByRelatedId() (success bool) {
	db := GetDB()
	defer CloseDB(db)
	if result := db.Where("r_id", ipct.RelatedId).First(ipct); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// DeleteByRelatedId 根据指定r_id删除指定记录
func (ipct *IpColorTag) DeleteByRelatedId() (success bool) {
	db := GetDB()
	defer CloseDB(db)
	if result := db.Where("r_id",ipct.RelatedId).Delete(ipct);result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}


// Update 更新指定ID的一条记录，列名和内容位于map中
func (ipct *IpColorTag) Update(updatedMap map[string]interface{}) (success bool) {
	updatedMap["update_datetime"] = time.Now()

	db := GetDB()
	defer CloseDB(db)
	if result := db.Model(ipct).Updates(updatedMap);result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}