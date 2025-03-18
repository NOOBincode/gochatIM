package mysql

import (
	"GochatIM/pkg/logger"
	"time"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// NewDB 创建数据库连接
func NewDB() (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(viper.GetString("mysql.host")+viper.GetString("mysql.port")), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名
		},
		Logger: logger.NewGormLogger(), // 使用自定义日志
	})
	if err != nil {
		return nil, err
	}

	// 获取底层的 SQL DB
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	logger.Info("数据库连接成功")
	return db, nil
}
