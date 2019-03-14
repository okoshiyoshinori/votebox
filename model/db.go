package model

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/okoshiyoshinori/votebox/config"
)

var db = newDBConn()

func newDBConn() *gorm.DB {
	config := config.GetDBServerConfig()
	connection := config.User + ":" + config.Pass + "@" + config.Protocol + "/" + config.DBname + "?" + config.Parsetime + "&" + config.Charset
	d, err := gorm.Open(config.Dbms, connection)
	if err != nil {
		panic(err.Error())
	}
	return d
}

func GetDBConn() *gorm.DB {
	return db
}
