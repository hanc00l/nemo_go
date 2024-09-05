package db

import "time"

type User struct {
	Id              int       `gorm:"primaryKey"`
	UserName        string    `gorm:"column:user_name"`
	UserPassword    string    `gorm:"column:user_password"`
	UserDescription string    `gorm:"column:user_description"`
	UserRole        string    `gorm:"column:user_role"`
	State           string    `gorm:"column:state"`
	SortOrder       int       `gorm:"column:sort_order"`
	CreateDatetime  time.Time `gorm:"column:create_datetime"`
	UpdateDatetime  time.Time `gorm:"column:update_datetime"`
}

// TableName 设置数据库关联的表名
func (*User) TableName() string {
	return "user"
}

// Get 根据ID查询记录
func (u *User) Get() (success bool) {
	db := GetDB()
	defer CloseDB(db)

	if result := db.First(u, u.Id); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// GetByUsername 根据username查询记录
func (u *User) GetByUsername() (success bool) {
	db := GetDB()
	defer CloseDB(db)

	if result := db.Where("user_name", u.UserName).First(u); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// Gets 根据指定的条件，查询满足要求的记录
func (u *User) Gets(searchMap map[string]interface{}, page, rowsPerPage int) (results []User, count int) {
	orderBy := "sort_order desc,user_name"

	db := GetDB()
	defer CloseDB(db)
	for column, value := range searchMap {
		db = db.Where(column, value)
	}
	db = db.Model(u)
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
func (u *User) Count(searchMap map[string]interface{}) (count int) {
	db := GetDB()
	defer CloseDB(db)

	for column, value := range searchMap {
		db = db.Where(column, value)
	}
	var result int64
	db.Model(u).Count(&result)

	return int(result)
}

// Add 插入一条新的记录，返回主键ID及成功标志
func (u *User) Add() (success bool) {
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

// Update 更新指定ID的一条记录，列名和内容位于map中
func (u *User) Update(updatedMap map[string]interface{}) (success bool) {
	updatedMap["update_datetime"] = time.Now()

	db := GetDB()
	defer CloseDB(db)
	if result := db.Model(u).Updates(updatedMap); result.RowsAffected == 1 {
		return true
	} else {
		return false
	}
}

// Delete 删除指定主键ID的一条记录
func (u *User) Delete() (success bool) {
	db := GetDB()
	defer CloseDB(db)
	if result := db.Delete(u, u.Id); result.RowsAffected == 1 {
		return true
	} else {
		return false
	}
}
