package model

import (
	"ggslayer/utils/config"
	"ggslayer/utils/mysql"
	"gorm.io/gorm"
)

var (
	db *gorm.DB
)

// Setup initializes the database instance
func Setup() {
	db = mysql.MysqlGet("ggslayer")
	if !config.IsDevEnv {
		//db.SetLogger(log.Loger)
	}

}

func CloseDB() {
	defer mysql.MysqlClose()
}

func GetDB() *gorm.DB {
	return db
}
