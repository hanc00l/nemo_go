package db

import (
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"time"
)

// 全局数据库连接
var globalDB *gorm.DB

// GetDB 获取一个数据库连接
func GetDB() *gorm.DB {
	if globalDB != nil {
		return globalDB
	}

	const RetriedNumber = 5
	const RetriedSleepTime = time.Second * 5
	RetriedCount := 0
	for {
		if RetriedCount > RetriedNumber {
			logging.RuntimeLog.Fatal("Failed to connect database")
			return nil
		}
		globalDB = getDB()
		if globalDB == nil {
			logging.RuntimeLog.Error("connect to database fail,retry...")
			RetriedCount++
			time.Sleep(RetriedSleepTime)
			continue
		}
		return globalDB
	}
}

func getDB() *gorm.DB {
	database := conf.GlobalServerConfig().Database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		database.Username, database.Password, database.Host, database.Port, database.Dbname)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return nil
	}
	//当没有开启debug模式的时候，gorm底层默认的log级别是Warn，
	//当SQL语句执行时间超过了100ms的时候就会触发Warn日志打印，同时错误的SQL语句也会触发。
	//设置为Silent后将不会显示任何SQL语句
	db.Logger = logger.Default.LogMode(logger.Silent)

	//设置连接池参数
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(10)
	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(100)
	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db
}

// CloseDB 显式关闭一个数据库连接
func CloseDB(db *gorm.DB) {
	//全局长连接模式不能关闭数据库连接
	//sql, _ := db.DB()
	//defer sql.Close()
}
