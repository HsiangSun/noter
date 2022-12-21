package orm

import (
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"noter/model"
	"noter/utils/config"
	"noter/utils/helper"
	"noter/utils/log"
)

var Gdb *gorm.DB

func InitDb() {
	sqlPath := fmt.Sprintf("%s%s", config.AppPath, "/db/note.db")
	log.Sugar.Infof("SqlPath:%s", sqlPath)
	db, err := gorm.Open(sqlite.Open(sqlPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Sugar.Errorf("open database err:%s", err)
	}
	err = db.AutoMigrate(model.Admin{}, model.Balance{}, model.Bill{}, model.Rate{})
	if err != nil {
		log.Sugar.Errorf("orm auto migrate have error:%s", err)
	}
	database, _ := db.DB()
	database.SetMaxOpenConns(2)
	err = database.Ping()
	if err != nil {
		log.Sugar.Errorf("db pring:%s", err)
	}
	Gdb = db

	LoadToMemory()
}

// 将记账员从数据库加载到内存中
func LoadToMemory() {

	rows, err := Gdb.Table("admins").Select("gid,uid").Rows()
	if err != nil {
		log.Sugar.Errorf("Load db data to memory has error:%s", err)
	}
	defer rows.Close()

	for rows.Next() {
		var gid string
		var uid string
		rows.Scan(&gid, &uid)
		helper.AddNorer(gid, uid)
	}

	log.Sugar.Infof("Load all db to memory")

}