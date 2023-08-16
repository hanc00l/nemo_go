package db

import (
	"time"
)

type Organization struct {
	Id             int       `gorm:"primaryKey"`
	OrgName        string    `gorm:"column:org_name"`
	Status         string    `gorm:"column:status"`
	SortOrder      int       `gorm:"column:sort_order" `
	WorkspaceId    int       `gorm:"column:workspace_id"`
	CreateDatetime time.Time `gorm:"column:create_datetime"`
	UpdateDatetime time.Time `gorm:"column:update_datetime" `
}

// TableName 设置数据库关联的表名
func (*Organization) TableName() string {
	return "organization"
}

// Get 查询指定主键ID的一条记录
func (org *Organization) Get() (success bool) {
	db := GetDB()
	defer CloseDB(db)
	if result := db.First(org, org.Id); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// Gets 根据查询条件执行数据库查询操作，返回查询结果数组
func (org *Organization) Gets(searchMap map[string]interface{}, page int, rowsPerPage int) (results []Organization) {
	orderBy := "sort_order desc,org_name"

	db := GetDB()
	defer CloseDB(db)
	for column, value := range searchMap {
		db = db.Where(column, value)
	}
	if rowsPerPage > 0 && page > 0 {
		db = db.Offset((page - 1) * rowsPerPage).Limit(rowsPerPage)
	}
	db.Model(org).Order(orderBy).Find(&results)
	return
}

// Add 插入一条新的记录，返回主键ID及成功标志
func (org *Organization) Add() (success bool) {
	org.CreateDatetime = time.Now()
	org.UpdateDatetime = time.Now()

	db := GetDB()
	defer CloseDB(db)
	if result := db.Create(org); result.RowsAffected == 1 {
		return true
	} else {
		return false
	}
}

// Update 更新指定ID的一条记录，列名和内容位于map中
func (org *Organization) Update(updatedMap map[string]interface{}) (success bool) {
	updatedMap["update_datetime"] = time.Now()

	db := GetDB()
	defer CloseDB(db)
	if result := db.Model(org).Updates(updatedMap); result.RowsAffected == 1 {
		return true
	} else {
		return false
	}
}

// Delete 删除指定主键ID的一条记录
func (org *Organization) Delete() (success bool) {
	db := GetDB()
	defer CloseDB(db)
	if result := db.Delete(org, org.Id); result.RowsAffected == 1 {
		return true
	} else {
		return false
	}
}

// Count 统计指定查询条件的记录数量
func (org *Organization) Count(searchMap map[string]interface{}) (count int) {
	db := GetDB()
	defer CloseDB(db)

	for column, value := range searchMap {
		db = db.Where(column, value)
	}
	var result int64
	db.Model(org).Count(&result)

	return int(result)
}
