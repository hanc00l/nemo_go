package db

import (
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"time"
)

type PortAttr struct {
	Id             int       `gorm:"primaryKey"`
	RelatedId      int       `gorm:"column:r_id"`
	Source         string    `gorm:"column:source"`
	Tag            string    `gorm:"column:tag"`
	Content        string    `gorm:"column:content"`
	Hash           string    `gorm:"column:hash"`
	CreateDatetime time.Time `gorm:"column:create_datetime"`
	UpdateDatetime time.Time `gorm:"column:update_datetime"`
}

func (*PortAttr) TableName() string {
	return "port_attr"
}

// Add 插入一条新的记录，返回主键ID及成功标志
func (portAttr *PortAttr) Add() (success bool) {
	portAttr.CreateDatetime = time.Now()
	portAttr.UpdateDatetime = time.Now()
	portAttr.Hash = utils.MD5(fmt.Sprintf("%d%s%s%s", portAttr.RelatedId, portAttr.Source, portAttr.Tag, portAttr.Content))

	db := GetDB()
	defer CloseDB(db)
	if result := db.Create(portAttr); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// GetByPortAttr 根据端口和属性查询一条记录
func (portAttr *PortAttr) GetByPortAttr() (success bool) {
	db := GetDB()
	defer CloseDB(db)

	hash := utils.MD5(fmt.Sprintf("%d%s%s%s", portAttr.RelatedId, portAttr.Source, portAttr.Tag, portAttr.Content))
	if result := db.Where("hash", hash).First(portAttr); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// GetsByRelatedId 根据查询条件执行数据库查询操作，返回查询结果数组
func (portAttr *PortAttr) GetsByRelatedId() (results []PortAttr) {
	orderBy := "tag,update_datetime desc"

	db := GetDB()
	defer CloseDB(db)
	db.Where("r_id", portAttr.RelatedId).Order(orderBy).Find(&results)
	return
}

// Update 更新指定ID的一条记录，列名和内容位于map中
func (portAttr *PortAttr) Update(updateMap map[string]interface{}) (success bool) {
	updateMap["update_datetime"] = time.Now()

	db := GetDB()
	defer CloseDB(db)
	if result := db.Model(portAttr).Updates(updateMap); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// Delete 删除指定主键ID的一条记录
func (portAttr *PortAttr) Delete() (success bool) {
	db := GetDB()
	defer CloseDB(db)
	if result := db.Delete(portAttr, portAttr.Id); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// SaveOrUpdate 保存、更新一条记录
func (portAttr *PortAttr) SaveOrUpdate() (success bool) {
	if portAttr.GetByPortAttr() {
		return portAttr.Update(map[string]interface{}{})
	} else {
		return portAttr.Add()
	}
}
