package db

import (
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"gorm.io/gorm"
	"net"
	"strings"
	"time"
)

type Ip struct {
	Id             int       `gorm:"primaryKey"`
	IpName         string    `gorm:"column:ip"`
	IpInt          uint64    `gorm:"column:ip_int"`
	OrgId          *int      `gorm:"column:org_id"` //使用指针可以处理数据库的NULL（go中传递nil）
	Location       string    `gorm:"column:location"`
	Status         string    `gorm:"column:status"`
	WorkspaceId    int       `gorm:"column:workspace_id"`
	PinIndex       int       `gorm:"column:pin_index"`
	CreateDatetime time.Time `gorm:"column:create_datetime"`
	UpdateDatetime time.Time `gorm:"column:update_datetime"`
}

// TableName 设置数据库关联的表名
func (*Ip) TableName() string {
	return "ip"
}

// Get 根据ID查询记录
func (ip *Ip) Get() (success bool) {
	db := GetDB()
	defer CloseDB(db)

	if result := db.First(ip, ip.Id); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// Add 插入一条新的记录
func (ip *Ip) Add() (success bool) {
	ip.CreateDatetime = time.Now()
	ip.UpdateDatetime = time.Now()
	if utils.CheckIPV4(ip.IpName) {
		ip.IpInt = uint64(utils.IPV4ToUInt32(ip.IpName))
	} else if utils.CheckIPV6(ip.IpName) {
		ip.IpInt = utils.IPV6Prefix64ToUInt64(ip.IpName)
		ip.IpName = utils.GetIPV6ParsedFormat(ip.IpName)
	}

	db := GetDB()
	defer CloseDB(db)
	if result := db.Create(ip); result.RowsAffected == 1 {
		return true
	} else {
		return false
	}
}

// GetByIp 根据IP查询记录
func (ip *Ip) GetByIp() (success bool) {
	db := GetDB()
	defer CloseDB(db)

	if ip.WorkspaceId > 0 {
		db = db.Where("workspace_id", ip.WorkspaceId)
	}
	if result := db.Where("ip", ip.IpName).First(ip); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// Update 更新指定ID的一条记录，列名和内容位于map中
func (ip *Ip) Update(updateMap map[string]interface{}) (success bool) {
	updateMap["update_datetime"] = time.Now()

	db := GetDB()
	defer CloseDB(db)
	if result := db.Model(ip).Updates(updateMap); result.RowsAffected == 1 {
		return true
	} else {
		return false
	}
}

// Delete 删除指定主键ID的一条记录
func (ip *Ip) Delete() (success bool) {
	db := GetDB()
	defer CloseDB(db)
	if result := db.Delete(ip, ip.Id); result.RowsAffected == 1 {
		return true
	} else {
		return false
	}
}

// Count 统计指定查询条件的记录数量
func (ip *Ip) Count(searchMap map[string]interface{}) (count int) {
	db := ip.makeWhere(searchMap).Model(ip)
	defer CloseDB(db)
	var result int64
	db.Count(&result)
	return int(result)
}

// Gets 根据指定的条件，查询满足要求的记录
func (ip *Ip) Gets(searchMap map[string]interface{}, page, rowsPerPage int, orderByDate bool) (results []Ip, count int) {
	orderByField := "ip_int"
	if orderByDate {
		orderByField = "update_datetime desc"
	}
	orderBy := "pin_index desc," + orderByField
	db := ip.makeWhere(searchMap).Model(ip)
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
func (ip *Ip) SaveOrUpdate() (success bool, isAdd bool) {
	oldRecord := &Ip{IpName: ip.IpName, WorkspaceId: ip.WorkspaceId}
	//如果记录已存在，则更新指定的字段
	if oldRecord.GetByIp() {
		updateMap := map[string]interface{}{}
		if ip.Status != "" {
			updateMap["status"] = ip.Status
		}
		if ip.OrgId != nil && *ip.OrgId != 0 {
			updateMap["org_id"] = ip.OrgId
		}
		if ip.Location != "" {
			updateMap["location"] = ip.Location
		}
		//更新记录
		ip.Id = oldRecord.Id
		return ip.Update(updateMap), false
	} else {
		//新增一条记录
		return ip.Add(), true
	}
}

// makeWhere 根据查询条件的不同的字段，组合生成count和search的查询条件
func (ip *Ip) makeWhere(searchMap map[string]interface{}) *gorm.DB {
	db := GetDB()
	//根据查询条件的不同的字段，组合生成查询条件
	for column, value := range searchMap {
		switch column {
		case "location":
			db = makeLike(value, column, db)
		case "domain":
			dbDomains := GetDB().Model(&Domain{}).Select("id").Where("domain like ?", fmt.Sprintf("%%%s%%", value))
			dbContent := GetDB().Model(&DomainAttr{}).Select("content").Where("tag='A' or tag='AAAA'").Where("r_id in (?)", dbDomains)
			db = db.Where("ip in (?)", dbContent)
			CloseDB(dbDomains)
			CloseDB(dbContent)
		case "ip":
			if utils.CheckIPV4(value.(string)) || utils.CheckIPV4Subnet(value.(string)) {
				_ip, _ipNet, err := net.ParseCIDR(value.(string))
				if err != nil {
					db = db.Where("ip", value)
				} else {
					ones, bits := _ipNet.Mask.Size()
					_ipStart := utils.IPV4ToUInt32(_ip.String())
					_ipEnd := _ipStart + (1 << (bits - ones)) - 1
					db = db.Where("ip_int between ? and ?", _ipStart, _ipEnd)
				}
			} else {
				// ipv6是模糊查询
				db = makeLike(value, column, db)
			}
		case "port":
			ports := strings.Split(value.(string), ",")
			dbPorts := GetDB().Model(&Port{}).Select("ip_id").Distinct("ip_id")
			for _, p := range ports {
				dbPorts = dbPorts.Or("port", p)
			}
			db = db.Where("id in (?)", dbPorts)
			CloseDB(dbPorts)
		case "port_status":
			portStatus := GetDB().Model(&Port{}).Select("ip_id").Where("status", value)
			db = db.Where("id in (?)", portStatus)
			CloseDB(portStatus)
		case "content":
			portAttr := GetDB().Model(&PortAttr{}).Select("r_id").Where("content like ?", fmt.Sprintf("%%%s%%", value))
			port := GetDB().Model(&Port{}).Select("ip_id").Where("id in (?)", portAttr)
			db = db.Where("id in (?)", port)
			CloseDB(portAttr)
			CloseDB(port)
		case "color_tag":
			colorTag := GetDB().Model(&IpColorTag{}).Select("r_id").Where("color", value)
			db = db.Where("id in (?)", colorTag)
			CloseDB(colorTag)
		case "memo_content":
			memoContent := GetDB().Model(&IpMemo{}).Select("r_id").Where("content like ?", fmt.Sprintf("%%%s%%", value))
			db = db.Where("id in (?)", memoContent)
			CloseDB(memoContent)
		case "date_delta":
			db = makeDateDelta(value.(int), "update_datetime", db)
		case "create_date_delta":
			daysToHour := 24 * value.(int)
			dayDelta, err := time.ParseDuration(fmt.Sprintf("-%dh", daysToHour))
			if err == nil {
				dbPorts := GetDB().Model(&Port{}).Select("ip_id").Distinct("ip_id").Where("create_datetime between ? and ?", time.Now().Add(dayDelta), time.Now())
				db = db.Where("id in (?)", dbPorts)
				CloseDB(dbPorts)
			}
		case "ip_http":
			http := GetDB().Model(&IpHttp{}).Select("r_id").Where("content like ?", fmt.Sprintf("%%%s%%", value))
			port := GetDB().Model(&Port{}).Select("ip_id").Where("id in (?)", http)
			db = db.Where("id in (?)", port)
			CloseDB(http)
			CloseDB(port)
		default:
			db = db.Where(column, value)
		}
	}
	return db
}
