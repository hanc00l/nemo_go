package db

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"time"
)

type RuntimeLog struct {
	Id             int       `gorm:"primaryKey"`
	Source         string    `gorm:"column:source"`
	File           string    `gorm:"column:file"`
	Func           string    `gorm:"column:func"`
	Level          string    `gorm:"column:level"`
	LevelInt       int       `gorm:"level_int"`
	Message        string    `gorm:"column:message"`
	CreateDatetime time.Time `gorm:"column:create_datetime"`
	UpdateDatetime time.Time `gorm:"column:update_datetime"`
}

func (RuntimeLog) TableName() string {
	return "runtimelog"
}

// Add 插入一条新的记录，返回主键ID及成功标志
func (l *RuntimeLog) Add() (success bool) {
	l.CreateDatetime = time.Now()
	l.UpdateDatetime = time.Now()
	/*
		var AllLevels = []Level{
			PanicLevel,
			FatalLevel,
			ErrorLevel,
			WarnLevel,
			InfoLevel,
			DebugLevel,
			TraceLevel,
		}
			// PanicLevel level, highest level of severity. Logs and then calls panic with the
			// message passed to Debug, Info, ...
			PanicLevel Level = iota
			// FatalLevel level. Logs and then calls `logger.Exit(1)`. It will exit even if the
			// logging level is set to Panic.
			FatalLevel
			// ErrorLevel level. Logs. Used for errors that should definitely be noted.
			// Commonly used for hooks to send errors to an error tracking service.
			ErrorLevel
			// WarnLevel level. Non-critical entries that deserve eyes.
			WarnLevel
			// InfoLevel level. General operational entries about what's going on inside the
			// application.
			InfoLevel
			// DebugLevel level. Usually only enabled when debugging. Very verbose logging.
			DebugLevel
			// TraceLevel level. Designates finer-grained informational events than the Debug.
			TraceLevel
	*/
	switch l.Level {
	case "panic":
		l.LevelInt = int(logrus.PanicLevel)
	case "fatal":
		l.LevelInt = int(logrus.FatalLevel)
	case "error":
		l.LevelInt = int(logrus.ErrorLevel)
	case "warning":
		l.LevelInt = int(logrus.WarnLevel)
	case "info":
		l.LevelInt = int(logrus.InfoLevel)
	case "debug":
		l.LevelInt = int(logrus.DebugLevel)
	default:
		l.LevelInt = 10
	}

	db := GetDB()
	defer CloseDB(db)
	if result := db.Create(l); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// Get 根据Id查询记录
func (l *RuntimeLog) Get() (success bool) {
	db := GetDB()
	defer CloseDB(db)

	if result := db.First(l, l.Id); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// Delete 删除指定主键ID的一条记录
func (l *RuntimeLog) Delete() (success bool) {
	db := GetDB()
	defer CloseDB(db)
	if result := db.Delete(l, l.Id); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// DeleteLogs 批量删除指定条件的记录
func (l *RuntimeLog) DeleteLogs(searchMap map[string]interface{}) (success bool) {
	db := l.makeWhere(searchMap).Model(l)
	defer CloseDB(db)

	if result := db.Delete(l); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// Count 统计指定查询条件的记录数量
func (l *RuntimeLog) Count(searchMap map[string]interface{}) (count int) {
	db := l.makeWhere(searchMap).Model(l)
	defer CloseDB(db)
	var result int64
	db.Count(&result)
	return int(result)
}

// makeWhere 根据查询条件的不同的字段，组合生成count和search的查询条件
func (l *RuntimeLog) makeWhere(searchMap map[string]interface{}) *gorm.DB {
	db := GetDB()
	//根据查询条件的不同的字段，组合生成查询条件
	for column, value := range searchMap {
		switch column {
		case "source":
			db = db.Where("source like ?", fmt.Sprintf("%%%s%%", value))
		case "file":
			db = db.Where("file like ?", fmt.Sprintf("%%%s%%", value))
		case "func":
			db = db.Where("func like ?", fmt.Sprintf("%%%s%%", value))
		case "level":
			db = db.Where("level", value)
		case "level_int":
			db = db.Where("level_int <= ?", value)
		case "message":
			db = db.Where("message like ?", fmt.Sprintf("%%%s%%", value))
		case "date_delta":
			daysToHour := 24 * value.(int)
			dayDelta, err := time.ParseDuration(fmt.Sprintf("-%dh", daysToHour))
			if err == nil {
				db = db.Where("update_datetime between ? and ?", time.Now().Add(dayDelta), time.Now())
			}
		default:
			db = db.Where(column, value)
		}
	}
	return db
}

// Gets 根据指定的条件，查询满足要求的记录
func (l *RuntimeLog) Gets(searchMap map[string]interface{}, page, rowsPerPage int) (results []RuntimeLog, count int) {
	orderBy := "update_datetime desc"

	db := l.makeWhere(searchMap).Model(l)
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
func (l *RuntimeLog) SaveOrUpdate() (success bool, isAdd bool) {
	return l.Add(), true
}
