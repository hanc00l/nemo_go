package db

import (
	"fmt"
	"gorm.io/gorm"
	"time"
)

type Domain struct {
	Id             int       `gorm:"primaryKey"`
	DomainName     string    `gorm:"column:domain"`
	OrgId          *int      `gorm:"column:org_id"` //使用指针可以处理数据库的NULL（go中传递nil）
	CreateDatetime time.Time `gorm:"column:create_datetime"`
	UpdateDatetime time.Time `gorm:"column:update_datetime"`
}

// TableName 设置数据库关联的表名
func (Domain) TableName() string {
	return "domain"
}

// Get 根据ID查询记录
func (domain *Domain) Get() (success bool) {
	db := GetDB()
	defer CloseDB(db)
	if result := db.First(domain, domain.Id); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// Add 插入一条新的记录
func (domain *Domain) Add() (success bool) {
	domain.CreateDatetime = time.Now()
	domain.UpdateDatetime = time.Now()

	db := GetDB()
	defer CloseDB(db)
	if result := db.Create(domain); result.RowsAffected == 1 {
		return true
	} else {
		return false
	}
}

// GetByDomain 根据IP查询记录
func (domain *Domain) GetByDomain() (success bool) {
	db := GetDB()
	defer CloseDB(db)
	if result := db.Where("domain = ?", domain.DomainName).First(domain); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// Update 更新指定ID的一条记录，列名和内容位于map中
func (domain *Domain) Update(updateMap map[string]interface{}) (success bool) {
	updateMap["update_datetime"] = time.Now()

	db := GetDB()
	defer CloseDB(db)
	if result := db.Model(domain).Updates(updateMap); result.RowsAffected == 1 {
		return true
	} else {
		return false
	}
}

// Delete 删除指定主键ID的一条记录
func (domain *Domain) Delete() (success bool) {
	db := GetDB()
	defer CloseDB(db)

	if result := db.Delete(domain, domain.Id); result.RowsAffected == 1 {
		return true
	} else {
		return false
	}
}

// Count 统计指定查询条件的记录数量
func (domain *Domain) Count(searchMap map[string]interface{}) (count int) {
	db := domain.makeWhere(searchMap).Model(domain)
	defer CloseDB(db)
	var result int64
	db.Count(&result)
	return int(result)
}

// makeWhere 根据查询条件的不同的字段，组合生成count和search的查询条件
func (domain *Domain) makeWhere(searchMap map[string]interface{}) *gorm.DB {
	db := GetDB()
	//根据查询条件的不同的字段，组合生成查询条件
	for column, value := range searchMap {
		switch column {
		case "org_id":
			db = db.Where("org_id", value)
		case "domain":
			db = db.Where("domain like ?", fmt.Sprintf("%%%s%%", value))
		case "ip":
			domainAttr := GetDB().Model(&DomainAttr{}).Select("r_id").Distinct("r_id").Where("tag", "A").Where("content", value)
			db = db.Where("id in (?)", domainAttr)
			CloseDB(domainAttr)
		case "color_tag":
			colorTag := GetDB().Model(&DomainColorTag{}).Select("r_id").Where("color", value)
			db = db.Where("id in (?)", colorTag)
			CloseDB(colorTag)
		case "memo_content":
			memoContent := GetDB().Model(&DomainMemo{}).Select("r_id").Where("content like ?", fmt.Sprintf("%%%s%%", value))
			db = db.Where("id in (?)", memoContent)
			CloseDB(memoContent)
		case "date_delta":
			daysToHour := 24 * value.(int)
			dayDelta, err := time.ParseDuration(fmt.Sprintf("-%dh", daysToHour))
			if err == nil {
				db = db.Where("update_datetime between ? and ?", time.Now().Add(dayDelta), time.Now())
			}
		case "create_date_delta":
			daysToHour := 24 * value.(int)
			dayDelta, err := time.ParseDuration(fmt.Sprintf("-%dh", daysToHour))
			if err == nil {
				db = db.Where("create_datetime between ? and ?", time.Now().Add(dayDelta), time.Now())
			}
		case "content":
			domainAttr := GetDB().Model(&DomainAttr{}).Select("r_id").Where("content like ?", fmt.Sprintf("%%%s%%", value))
			db = db.Where("id in (?)", domainAttr)
			CloseDB(domainAttr)
		default:
			db = db.Where(column, value)
		}
	}
	return db
}

// Gets 根据指定的条件，查询满足要求的记录
func (domain *Domain) Gets(searchMap map[string]interface{}, page, rowsPerPage int, orderByDate bool) (results []Domain, count int) {
	orderBy := "domain"
	if orderByDate {
		orderBy = "update_datetime desc"
	}

	db := domain.makeWhere(searchMap).Model(domain)
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

// SaveOrUpdate 保存、更新一条记录
func (domain *Domain) SaveOrUpdate() (success bool) {
	oldRecord := &Domain{DomainName: domain.DomainName}
	//如果记录已存在，则更新指定的字段
	if oldRecord.GetByDomain() {
		updateMap := make(map[string]interface{})
		if domain.OrgId != nil && *domain.OrgId != 0 {
			updateMap["org_id"] = domain.OrgId
		}
		//更新记录
		domain.Id = oldRecord.Id
		return domain.Update(updateMap)
	} else {
		//新增一条记录
		return domain.Add()
	}
}
