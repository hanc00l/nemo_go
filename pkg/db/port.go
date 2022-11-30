package db

import "time"

type Port struct {
	Id             int       `gorm:"primaryKey"`
	IpId           int       `gorm:"column:ip_id"'`
	PortNum        int       `gorm:"column:port"`
	Status         string    `gorm:"column:status"`
	CreateDatetime time.Time `gorm:"column:create_datetime"`
	UpdateDatetime time.Time `gorm:"column:update_datetime"`
}

func (Port) TableName() string {
	return "port"
}

// Add 插入一条新的记录，返回主键ID及成功标志
func (port *Port) Add() (success bool) {
	port.CreateDatetime = time.Now()
	port.UpdateDatetime = time.Now()

	db := GetDB()
	defer CloseDB(db)
	if result := db.Create(port); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// GetByIPPort 根据IP和Port查询指定记录
func (port *Port) GetByIPPort() (success bool) {
	db := GetDB()
	defer CloseDB(db)

	if result := db.Where("ip_id", port.IpId).Where("port", port.PortNum).First(port); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// GetsByIPId 根据ip_id返回所有的端口记录
func (port *Port) GetsByIPId() (results []Port) {
	orderBy := "port"

	db := GetDB()
	defer CloseDB(db)
	db.Order(orderBy).Where("ip_id", port.IpId).Find(&results)
	return
}

// Update 更新指定ID的一条记录，列名和内容位于map中
func (port *Port) Update(updatedMap map[string]interface{}) (success bool) {
	updatedMap["update_datetime"] = time.Now()

	db := GetDB()
	defer CloseDB(db)
	if result := db.Model(port).Updates(updatedMap); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// SaveOrUpdate 保存、更新一条记录
func (port *Port) SaveOrUpdate() (success bool, isNew bool) {
	oldRecord := &Port{IpId: port.IpId, PortNum: port.PortNum}
	if oldRecord.GetByIPPort() {
		updateMap := map[string]interface{}{}
		if port.Status != "" {
			updateMap["status"] = port.Status
		}
		port.Id = oldRecord.Id
		return port.Update(updateMap), false
	} else {
		return port.Add(), true
	}
}
