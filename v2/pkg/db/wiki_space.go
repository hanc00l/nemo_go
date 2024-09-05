package db

import "time"

type WikiSpace struct {
	Id             int       `gorm:"primaryKey"`
	WorkspaceId    int       `gorm:"column:workspace_id"`
	WikiSpaceId    string    `gorm:"column:wiki_space_id"`
	CreateDatetime time.Time `gorm:"column:create_datetime"`
	UpdateDatetime time.Time `gorm:"column:update_datetime"`
}

func (*WikiSpace) TableName() string {
	return "wiki_space"
}

// Get 根据ID查询记录
func (w *WikiSpace) Get() (success bool) {
	db := GetDB()
	defer CloseDB(db)

	if result := db.First(w, w.Id); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// GetByWorkspaceId 根据workspace_id查询记录
func (w *WikiSpace) GetByWorkspaceId() (success bool) {
	db := GetDB()
	defer CloseDB(db)

	if result := db.Where("workspace_id", w.WorkspaceId).First(w); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// Add 插入一条新的记录，返回主键ID及成功标志
func (w *WikiSpace) Add() (success bool) {
	w.CreateDatetime = time.Now()
	w.UpdateDatetime = time.Now()

	db := GetDB()
	defer CloseDB(db)
	if result := db.Create(w); result.RowsAffected == 1 {
		return true
	} else {
		return false
	}
}

// Update 更新指定ID的一条记录，列名和内容位于map中
func (w *WikiSpace) Update(updatedMap map[string]interface{}) (success bool) {
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
func (w *WikiSpace) Delete() (success bool) {
	db := GetDB()
	defer CloseDB(db)
	if result := db.Delete(w, w.Id); result.RowsAffected == 1 {
		return true
	} else {
		return false
	}
}
