package manager

import (
	"database/sql"
	"fmt"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Manager struct {
	db    *gorm.DB
	redis redis.UniversalClient
}

var m *Manager

func Init(sqlDB *sql.DB, redisClient redis.UniversalClient) {
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	if err != nil {
		panic(fmt.Errorf("init gorm with postgres failed: %w", err))
	}
	m = &Manager{
		db:    gormDB,
		redis: redisClient,
	}
}
