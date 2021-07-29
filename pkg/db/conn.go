package db

import (
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// GetDB 获取一个数据库连接
func GetDB() *gorm.DB {
	user := conf.Nemo.Database.Username
	pass := conf.Nemo.Database.Password
	host := conf.Nemo.Database.Host
	port := conf.Nemo.Database.Port
	dbname := conf.Nemo.Database.Dbname
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, pass, host, port, dbname)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		logging.RuntimeLog.Fatal("Connect to database fail!")
		return nil
	}
	//当没有开启debug模式的时候，gorm底层默认的log级别是Warn，
	//当SQL语句执行时间超过了100ms的时候就会触发Warn日志打印，同时错误的SQL语句也会触发。
	//设置为Silent后将不会显示任何SQL语句
	db.Logger = logger.Default.LogMode(logger.Silent)

	return db
}

// CloseDB 显式关闭一个数据库连接
func CloseDB(db *gorm.DB) {
	sql, _ := db.DB()
	defer sql.Close()
}
