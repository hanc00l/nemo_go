package db

import "time"

type DomainColorTag struct {
	Id             int       `gorm:"primaryKey"`
	RelatedId      int       `gorm:"column:r_id"`
	Color          string    `gorm:"column:color"`
	CreateDatetime time.Time `gorm:"column:create_datetime"`
	UpdateDatetime time.Time `gorm:"column:update_datetime"`
}

func (DomainColorTag) TableName() string {
	return "domain_color_tag"
}

//Add 插入一条新的记录
func (dct *DomainColorTag) Add() (success bool) {
	dct.CreateDatetime = time.Now()
	dct.UpdateDatetime = time.Now()

	db := GetDB()
	defer CloseDB(db)
	if result := db.Create(dct); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// GetByRelatedId 根据指定r_id查询一条记录
func (dct *DomainColorTag) GetByRelatedId() (success bool) {
	db := GetDB()
	defer CloseDB(db)

	if result := db.Where("r_id", dct.RelatedId).First(dct); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// DeleteByRelatedId 根据指定r_id删除指定记录
func (dct *DomainColorTag) DeleteByRelatedId() (success bool) {
	db := GetDB()
	defer CloseDB(db)
	if result := db.Where("r_id", dct.RelatedId).Delete(dct); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// Update 更新指定ID的一条记录，列名和内容位于map中
func (dct *DomainColorTag) Update(updatedMap map[string]interface{}) (success bool) {
	updatedMap["update_datetime"] = time.Now()

	db := GetDB()
	defer CloseDB(db)
	if result := db.Model(dct).Updates(updatedMap); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}
