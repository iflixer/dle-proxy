package database

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"os"
	"strconv"
)

type Service struct {
	DB *gorm.DB
}

func NewService(mysqlURL string) (s *Service, err error) {

	db, err := gorm.Open(mysql.Open(mysqlURL), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	debugMode := os.Getenv("MYSQL_DEBUG_MODE")
	if debugMode != "" {
		d, _ := strconv.Atoi(debugMode)
		db.Logger = logger.Default.LogMode(logger.LogLevel(d))
	}

	s = &Service{
		DB: db,
	}

	return
}
