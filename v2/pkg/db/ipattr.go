package db

import (
	"fmt"
	"github.com/hanc00l/nemo_go/v2/pkg/utils"
	"time"
)

type IpAttr struct {
	Id             int       `gorm:"primaryKey"`
	RelatedId      int       `gorm:"column:r_id"`
	Source         string    `gorm:"column:source"`
	Tag            string    `gorm:"column:tag"`
	Content        string    `gorm:"column:content"`
	Hash           string    `gorm:"column:hash"`
	CreateDatetime time.Time `gorm:"column:create_datetime"`
	UpdateDatetime time.Time `gorm:"column:update_datetime"`
}

func (*IpAttr) TableName() string {
	return "ip_attr"
}

// Add 插入一条新的记录，返回主键ID及成功标志
func (ipAttr *IpAttr) Add() (success bool) {
	ipAttr.CreateDatetime = time.Now()
	ipAttr.UpdateDatetime = time.Now()
	ipAttr.Hash = utils.MD5(fmt.Sprintf("%d%s%s%s", ipAttr.RelatedId, ipAttr.Source, ipAttr.Tag, ipAttr.Content))

	db := GetDB()
	defer CloseDB(db)
	if result := db.Create(ipAttr); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// Get 查询指定主键ID的一条记录
func (ipAttr *IpAttr) Get() (success bool) {
	db := GetDB()
	defer CloseDB(db)

	if result := db.First(ipAttr, ipAttr.Id); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// Gets 根据查询条件执行数据库查询操作，返回查询结果数组
func (ipAttr *IpAttr) Gets(searchMap map[string]interface{}, page int, rowsPerPage int) (results []IpAttr) {
	orderBy := "tag,update_datetime desc"

	db := GetDB()
	defer CloseDB(db)
	for column, value := range searchMap {
		db = db.Where(column, value)
	}
	if rowsPerPage > 0 && page > 0 {
		db = db.Offset((page - 1) * rowsPerPage).Limit(rowsPerPage)
	}
	db.Order(orderBy).Find(&results)
	return results
}

// GetsByRelatedId 根据查询条件执行数据库查询操作，返回查询结果数组
func (ipAttr *IpAttr) GetsByRelatedId() (results []IpAttr) {
	searchMap := map[string]interface{}{}
	searchMap["r_id"] = ipAttr.RelatedId

	return ipAttr.Gets(searchMap, 0, 0)
}

// Update 更新指定ID的一条记录，列名和内容位于map中
func (ipAttr *IpAttr) Update(updatedMap map[string]interface{}) (success bool) {
	updatedMap["update_datetime"] = time.Now()
	updatedMap["hash"] = utils.MD5(fmt.Sprintf("%d%s%s%s", ipAttr.RelatedId, ipAttr.Source, ipAttr.Tag, ipAttr.Content))

	db := GetDB()
	defer CloseDB(db)
	if result := db.Model(&ipAttr).Updates(updatedMap); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// Delete 删除指定主键ID的一条记录
func (ipAttr *IpAttr) Delete() (success bool) {
	db := GetDB()
	defer CloseDB(db)

	result := db.Delete(ipAttr, ipAttr.Id)
	if result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}
