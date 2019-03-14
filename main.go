package main

import (
	"io"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/okoshiyoshinori/votebox/boxwebsock"
	"github.com/okoshiyoshinori/votebox/config"
	"github.com/okoshiyoshinori/votebox/logger"
	"github.com/okoshiyoshinori/votebox/model"
	"github.com/okoshiyoshinori/votebox/router"
	"github.com/okoshiyoshinori/votebox/textimage"
)

var (
	db        *gorm.DB
	r         *gin.Engine
	logfile   *os.File
	errorfile *os.File
)

func init() {
	logfile, _ = os.Create(config.APConfig.Log)
	gin.DefaultWriter = io.MultiWriter(logfile)
	errorfile, _ = os.Create(config.APConfig.ErrorLog)
	gin.DefaultErrorWriter = io.MultiWriter(errorfile)
	logger.Newlogger(logfile, errorfile)
	gin.SetMode(gin.ReleaseMode)
	db = model.GetDBConn()
	r = router.GetRouter()
}

func main() {
	defer func() {
		db.Close()
		logfile.Close()
		errorfile.Close()
	}()
	go boxwebsock.RootHub.Run()
	go textimage.Sub.Run()
	go boxwebsock.Dworker.Run()
	go boxwebsock.Receive()
	r.Run(config.APConfig.Port)
}
