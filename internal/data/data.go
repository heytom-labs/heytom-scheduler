package data

import (
	"heytom-scheduler/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewTaskRepo, NewExecutionRepo)

// Data .
type Data struct {
	db *gorm.DB
}

// NewData .
func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	log := log.NewHelper(logger)
	
	// 初始化数据库连接
	db, err := gorm.Open(mysql.Open(c.Database.Source), &gorm.Config{})
	if err != nil {
		log.Errorf("failed to connect database: %v", err)
		return nil, nil, err
	}
	
	// 自动迁移表结构
	if err := db.AutoMigrate(&Task{}, &TaskExecution{}); err != nil {
		log.Errorf("failed to migrate database: %v", err)
		return nil, nil, err
	}
	
	data := &Data{
		db: db,
	}
	
	cleanup := func() {
		log.Info("closing the data resources")
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}
	
	return data, cleanup, nil
}
