package db

import (
	"time"
)

type IpHttp struct {
	Id             int       `gorm:"primaryKey"`
	RelatedId      int       `gorm:"column:r_id"`
	Source         string    `gorm:"column:source"`
	Tag            string    `gorm:"column:tag"`
	Content        string    `gorm:"column:content"`
	CreateDatetime time.Time `gorm:"column:create_datetime"`
	UpdateDatetime time.Time `gorm:"column:update_datetime"`
}

func (IpHttp) TableName() string {
	return "ip_http"
}

// Add 插入一条新的记录，返回主键ID及成功标志
func (i *IpHttp) Add() (success bool) {
	i.CreateDatetime = time.Now()
	i.UpdateDatetime = time.Now()

	db := GetDB()
	defer CloseDB(db)
	if result := db.Create(i); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// GetsByRelatedId 根据查询条件执行数据库查询操作，返回查询结果数组
func (i *IpHttp) GetsByRelatedId() (results []IpHttp) {
	orderBy := "tag,update_datetime desc"

	db := GetDB()
	defer CloseDB(db)
	db.Where("r_id", i.RelatedId).Order(orderBy).Find(&results)
	return
}

// GetByRelatedIdAndTag 根据端口、tag查询一条记录
func (i *IpHttp) GetByRelatedIdAndTag() (success bool) {
	db := GetDB()
	defer CloseDB(db)

	if result := db.Where("r_id", i.RelatedId).Where("tag", i.Tag).First(i); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// Update 更新指定ID的一条记录，列名和内容位于map中
func (i *IpHttp) Update(updateMap map[string]interface{}) (success bool) {
	updateMap["update_datetime"] = time.Now()

	db := GetDB()
	defer CloseDB(db)
	if result := db.Model(i).Updates(updateMap); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// Delete 删除指定主键ID的一条记录
func (i *IpHttp) Delete() (success bool) {
	db := GetDB()
	defer CloseDB(db)
	if result := db.Delete(i, i.Id); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// SaveOrUpdate 保存、更新一条记录
func (i *IpHttp) SaveOrUpdate() (success bool) {
	oldRecord := &IpHttp{RelatedId: i.RelatedId, Tag: i.Tag}
	if oldRecord.GetByRelatedIdAndTag() {
		updateMap := make(map[string]interface{})
		if i.Source != "" {
			updateMap["source"] = i.Source
		}
		if i.Content != "" {
			updateMap["content"] = i.Content
		}
		//更新记录
		i.Id = oldRecord.Id
		return i.Update(updateMap)
	} else {
		return i.Add()
	}
}
