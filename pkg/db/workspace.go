package db

import (
	"github.com/google/uuid"
	"time"
)

type Workspace struct {
	Id                   int       `gorm:"primaryKey"`
	WorkspaceName        string    `gorm:"column:workspace_name"`
	WorkspaceGUID        string    `gorm:"column:workspace_guid"`
	WorkspaceDescription string    `gorm:"column:workspace_description"`
	State                string    `gorm:"column:state"`
	SortOrder            int       `gorm:"column:sort_order"`
	CreateDatetime       time.Time `gorm:"column:create_datetime"`
	UpdateDatetime       time.Time `gorm:"column:update_datetime"`
}

func (Workspace) TableName() string {
	return "workspace"
}

// Get 根据ID查询记录
func (w *Workspace) Get() (success bool) {
	db := GetDB()
	defer CloseDB(db)

	if result := db.First(w, w.Id); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// GetByGUID 根据guid查询记录
func (w *Workspace) GetByGUID() (success bool) {
	db := GetDB()
	defer CloseDB(db)

	if result := db.Where("workspace_guid", w.WorkspaceGUID).First(w); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// Gets 根据指定的条件，查询满足要求的记录
func (w *Workspace) Gets(searchMap map[string]interface{}, page, rowsPerPage int) (results []Workspace, count int) {
	orderBy := "sort_order desc,workspace_name"

	db := GetDB()
	defer CloseDB(db)
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

// Count 统计指定查询条件的记录数量
func (w *Workspace) Count(searchMap map[string]interface{}) (count int) {
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
func (w *Workspace) Add() (success bool) {
	w.CreateDatetime = time.Now()
	w.UpdateDatetime = time.Now()
	w.WorkspaceGUID = uuid.New().String()

	db := GetDB()
	defer CloseDB(db)
	if result := db.Create(w); result.RowsAffected == 1 {
		return true
	} else {
		return false
	}
}

// Update 更新指定ID的一条记录，列名和内容位于map中
func (w *Workspace) Update(updatedMap map[string]interface{}) (success bool) {
	updatedMap["update_datetime"] = time.Now()

	db := GetDB()
	defer CloseDB(db)
	if result := db.Model(w).Updates(updatedMap); result.RowsAffected == 1 {
		return true
	} else {
		return false
	}
}

// Delete 删除指定主键ID的一条记录
func (w *Workspace) Delete() (success bool) {
	db := GetDB()
	defer CloseDB(db)
	if result := db.Delete(w, w.Id); result.RowsAffected == 1 {
		return true
	} else {
		return false
	}
}
