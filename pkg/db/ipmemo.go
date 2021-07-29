package db

import "time"

type IpMemo struct {
	Id             int       `gorm:"primaryKey"`
	RelatedId      int       `gorm:"column:r_id"`
	Content        string    `gorm:"column:content"`
	CreateDatetime time.Time `gorm:"column:create_datetime"`
	UpdateDatetime time.Time `gorm:"column:update_datetime"`
}

func (IpMemo) TableName() string {
	return "ip_memo"
}

//Add 插入一条新的记录
func (ipmemo *IpMemo) Add() (success bool) {
	ipmemo.CreateDatetime = time.Now()
	ipmemo.UpdateDatetime = time.Now()

	db := GetDB()
	defer CloseDB(db)
	if result := db.Create(ipmemo);result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// GetByRelatedId 根据指定r_id查询一条记录
func (ipmemo *IpMemo) GetByRelatedId() (success bool) {
	db := GetDB()
	defer CloseDB(db)
	if result := db.Where("r_id", ipmemo.RelatedId).First(ipmemo); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// DeleteByRelatedId 根据指定r_id删除指定记录
func (ipmemo *IpMemo) DeleteByRelatedId() (success bool) {
	db := GetDB()
	defer CloseDB(db)
	if result := db.Where("r_id",ipmemo.RelatedId).Delete(ipmemo);result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// Update 更新指定ID的一条记录，列名和内容位于map中
func (ipmemo *IpMemo) Update(updateMap map[string]interface{}) (success bool) {
	updateMap["update_datetime"] = time.Now()

	db := GetDB()
	defer CloseDB(db)
	if result := db.Model(ipmemo).Updates(updateMap);result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}
