package bootstrap

import (
	"fmt"
	"noter/telgram"
	"noter/utils/config"
	"noter/utils/log"
	"noter/utils/orm"
	"os"
	"os/signal"
	"syscall"
)

func init() {
	config.InitConfig()

	fmt.Println("PATH!!!!!!" + config.AppPath)

	//init log
	log.InitLog()

	orm.InitDb()
	log.Sugar.Info("****************Bot running*********************")

	go func() {
		defer func() {
			if err := recover(); err != nil {
				//log.Printf("server bot err:%+v \n", err)
				log.Sugar.Error("****************Bot have error will restart*********************")
			}
		}()
		telgram.BootRun()
	}()

	//kill program
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

}
