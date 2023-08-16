package db

import "time"

type UserWorkspace struct {
	Id             int       `gorm:"primaryKey"`
	UserId         int       `gorm:"column:user_id"`
	WorkspaceId    int       `gorm:"column:workspace_id"`
	CreateDatetime time.Time `gorm:"column:create_datetime"`
	UpdateDatetime time.Time `gorm:"column:update_datetime"`
}

func (*UserWorkspace) TableName() string {
	return "user_workspace"
}

// Get 根据ID查询记录
func (u *UserWorkspace) Get() (success bool) {
	db := GetDB()
	defer CloseDB(db)

	if result := db.First(u, u.Id); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// GetsByUserId 根据userId获取workspace
func (u *UserWorkspace) GetsByUserId(userId int) (results []UserWorkspace) {
	orderBy := "update_datetime desc"

	db := GetDB()
	defer CloseDB(db)
	db.Order(orderBy).Where("user_id", userId).Find(&results)

	return
}

// GetByUserAndWorkspaceId 根据userId、workspaceId获取记录
func (u *UserWorkspace) GetByUserAndWorkspaceId() bool {
	db := GetDB()
	defer CloseDB(db)

	if result := db.Where("user_id", u.UserId).Where("workspace_id", u.WorkspaceId).First(u); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// Add 插入一条新的记录，返回主键ID及成功标志
func (u *UserWorkspace) Add() (success bool) {
	u.CreateDatetime = time.Now()
	u.UpdateDatetime = time.Now()

	db := GetDB()
	defer CloseDB(db)
	if result := db.Create(u); result.RowsAffected == 1 {
		return true
	} else {
		return false
	}
}

// RemoveUserWorkspace 删除指定用户的全部工作空间
func (u *UserWorkspace) RemoveUserWorkspace(userId int) {
	db := GetDB()
	defer CloseDB(db)

	db.Where("user_id", userId).Delete(u)
}
