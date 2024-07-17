package db

import (
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"time"
)

const AttrContentSize = 4000

type DomainAttr struct {
	Id             int       `gorm:"primaryKey"`
	RelatedId      int       `gorm:"column:r_id"`
	Source         string    `gorm:"column:source"`
	Tag            string    `gorm:"column:tag"`
	Content        string    `gorm:"column:content"`
	Hash           string    `gorm:"column:hash"`
	CreateDatetime time.Time `gorm:"column:create_datetime"`
	UpdateDatetime time.Time `gorm:"column:update_datetime"`
}

func (*DomainAttr) TableName() string {
	return "domain_attr"
}

// Add 插入一条新的记录，返回主键ID及成功标志
func (domainAttr *DomainAttr) Add() (success bool) {
	domainAttr.CreateDatetime = time.Now()
	domainAttr.UpdateDatetime = time.Now()
	domainAttr.Hash = utils.MD5(fmt.Sprintf("%d%s%s%s", domainAttr.RelatedId, domainAttr.Source, domainAttr.Tag, domainAttr.Content))

	db := GetDB()
	defer CloseDB(db)
	if result := db.Create(domainAttr); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// GetByDomainAttr 根据域名的属性查询一条记录
func (domainAttr *DomainAttr) GetByDomainAttr() (success bool) {
	db := GetDB()
	defer CloseDB(db)

	hash := utils.MD5(fmt.Sprintf("%d%s%s%s", domainAttr.RelatedId, domainAttr.Source, domainAttr.Tag, domainAttr.Content))
	if result := db.Where("hash", hash).First(domainAttr); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// GetsByRelatedId 根据查询条件执行数据库查询操作，返回查询结果数组
func (domainAttr *DomainAttr) GetsByRelatedId() (results []DomainAttr) {
	db := GetDB()
	defer CloseDB(db)

	db.Where("r_id", domainAttr.RelatedId).Order("tag,update_datetime desc").Find(&results)
	return
}

// GetsByRelatedIdByDateAsc 根据查询条件执行数据库查询操作，返回查询结果数组；按日期升序排列
func (domainAttr *DomainAttr) GetsByRelatedIdByDateAsc() (results []DomainAttr) {
	db := GetDB()
	defer CloseDB(db)

	db.Where("r_id", domainAttr.RelatedId).Order("tag,update_datetime asc").Find(&results)
	return
}

// Update 更新指定ID的一条记录，列名和内容位于map中
func (domainAttr *DomainAttr) Update(updatedMap map[string]interface{}) (success bool) {
	updatedMap["update_datetime"] = time.Now()

	db := GetDB()
	defer CloseDB(db)
	if result := db.Model(&domainAttr).Updates(updatedMap); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// Delete 删除指定主键ID的一条记录
func (domainAttr *DomainAttr) Delete() (success bool) {
	db := GetDB()
	defer CloseDB(db)

	result := db.Delete(domainAttr, domainAttr.Id)
	if result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// DeleteByRelatedIDAndSource 删除RID和source删除相应的属性
func (domainAttr *DomainAttr) DeleteByRelatedIDAndSource() (success bool) {
	db := GetDB()
	defer CloseDB(db)

	result := db.Where("r_id", domainAttr.RelatedId).Where("source", domainAttr.Source).Delete(domainAttr)
	if result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// SaveOrUpdate 保存、更新一条记录
func (domainAttr *DomainAttr) SaveOrUpdate() bool {
	if domainAttr.GetByDomainAttr() {
		return domainAttr.Update(map[string]interface{}{})
	} else {
		return domainAttr.Add()
	}
}
