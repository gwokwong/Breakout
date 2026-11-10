package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// safeInt 安全地将接口值转为int
func safeInt(v interface{}, defaultVal int) int {
	switch val := v.(type) {
	case int:
		return val
	case int64:
		return int(val)
	case float64:
		return int(val)
	case string:
		if n, err := strconv.Atoi(val); err == nil {
			return n
		}
	}
	return defaultVal
}

var (
	DB *gorm.DB
)

// Init 初始化服务
func Init() {
	// 初始化数据库
	initDB()
	
	// 初始化中间件
	InitMiddleware()
}

// initDB 初始化数据库连接
func initDB() {
	config := g.Cfg().MustGet(context.Background(), "database").Map()
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		config["user"],
		config["password"],
		config["host"],
		safeInt(config["port"], 3306),
		config["name"],
		config["charset"],
	)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{SingularTable: true},
	})
	if err != nil {
		g.Log().Fatal(context.Background(), "数据库连接失败:", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		g.Log().Fatal(context.Background(), "获取数据库实例失败:", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(safeInt(config["maxIdleConns"], 10))
	sqlDB.SetMaxOpenConns(safeInt(config["maxOpenConns"], 100))
	sqlDB.SetConnMaxLifetime(time.Duration(safeInt(config["connMaxLifetime"], 3600)) * time.Second)

	g.Log().Info(context.Background(), "数据库连接成功")
}