package db

import (
	"fmt"
	"gorm.io/gorm"
	"time"
)

type TaskMain struct {
	Id              int        `gorm:"primaryKey"`
	TaskId          string     `gorm:"column:task_id"`
	TaskName        string     `gorm:"column:task_name"`
	KwArgs          string     `gorm:"column:kwargs"`
	State           string     `gorm:"column:state"`
	Result          string     `gorm:"column:result"`
	ReceivedTime    time.Time  `gorm:"column:received"`
	StartedTime     *time.Time `gorm:"column:started"`
	SucceededTime   *time.Time `gorm:"column:succeeded"`
	ProgressMessage string     `gorm:"column:progress_message"`
	CronTaskId      string     `gorm:"column:cron_id"`
	WorkspaceId     int        `gorm:"column:workspace_id"`
	CreateDatetime  time.Time  `gorm:"column:create_datetime"`
	UpdateDatetime  time.Time  `gorm:"column:update_datetime"`
}

func (TaskMain) TableName() string {
	return "task_main"
}

// Add 插入一条新的记录，返回主键ID及成功标志
func (t *TaskMain) Add() (success bool) {
	now := time.Now()
	t.ReceivedTime = now
	t.CreateDatetime = now
	t.UpdateDatetime = now

	db := GetDB()
	defer CloseDB(db)
	if result := db.Create(t); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// Get 根据ID查询记录
func (t *TaskMain) Get() (success bool) {
	db := GetDB()
	defer CloseDB(db)

	if result := db.First(t, t.Id); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// GetByTaskId 根据TaskID（不是数据库ID）精确查询一条记录
func (t *TaskMain) GetByTaskId() (success bool) {
	db := GetDB()
	defer CloseDB(db)
	if result := db.Where("task_id", t.TaskId).First(t); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// Update 更新指定ID的一条记录，列名和内容位于map中
func (t *TaskMain) Update(updateMap map[string]interface{}) (success bool) {
	updateMap["update_datetime"] = time.Now()

	db := GetDB()
	defer CloseDB(db)
	if result := db.Model(t).Updates(updateMap); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// Delete 删除指定主键ID的一条记录
func (t *TaskMain) Delete() (success bool) {
	db := GetDB()
	defer CloseDB(db)
	if result := db.Delete(t, t.Id); result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

// Count 统计指定查询条件的记录数量
func (t *TaskMain) Count(searchMap map[string]interface{}) (count int) {
	db := t.makeWhere(searchMap).Model(t)
	defer CloseDB(db)
	var result int64
	db.Count(&result)
	return int(result)
}

// makeWhere 根据查询条件的不同的字段，组合生成count和search的查询条件
func (t *TaskMain) makeWhere(searchMap map[string]interface{}) *gorm.DB {
	db := GetDB()
	for column, value := range searchMap {
		switch column {
		case "task_name":
			db = db.Where("task_name like ?", fmt.Sprintf("%%%s%%", value))
		case "kwargs":
			db = db.Where("kwargs like ?", fmt.Sprintf("%%%s%%", value))
		case "result":
			db = db.Where("result like ?", fmt.Sprintf("%%%s%%", value))
		case "state":
			db = db.Where("state", value)
		case "date_delta":
			daysToHour := 24 * value.(int)
			dayDelta, err := time.ParseDuration(fmt.Sprintf("-%dh", daysToHour))
			if err == nil {
				db = db.Where("update_datetime between ? and ?", time.Now().Add(dayDelta), time.Now())
			}
		case "cron_id":
			db = db.Where("cron_id", value)
		case "workspace_id":
			db = db.Where("workspace_id", value)
		default:
			db = db.Where(column, value)
		}
	}
	return db
}

// Gets 根据指定的条件，查询满足要求的记录
func (t *TaskMain) Gets(searchMap map[string]interface{}, page, rowsPerPage int) (results []TaskMain, count int) {
	orderBy := "update_datetime desc"

	db := t.makeWhere(searchMap).Model(t)
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
func (t *TaskMain) SaveOrUpdate() (success bool) {
	oldRecord := &TaskRun{TaskId: t.TaskId}
	if oldRecord.GetByTaskId() {
		updateMap := map[string]interface{}{}
		if t.State != "" {
			updateMap["state"] = t.State
		}
		if t.Result != "" {
			updateMap["result"] = t.Result
		}
		if t.ProgressMessage != "" {
			updateMap["progress_message"] = t.ProgressMessage
		}
		t.Id = oldRecord.Id
		return t.Update(updateMap)
	} else {
		return t.Add()
	}
}
