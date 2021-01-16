package tests

import (
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/wujie1993/waves/pkg/db"
	"github.com/wujie1993/waves/pkg/setting"
)

var (
	ServiceEndpoint string
	EtcdEndpoint    string
)

func init() {
	if ServiceEndpoint == "" {
		ServiceEndpoint = "http://localhost:8000/deployer"
	}
	if EtcdEndpoint == "" {
		EtcdEndpoint = "localhost:2379"
	}

	initLog()
	initDB()
}

func initLog() {
	log.SetOutput(os.Stdout)
	log.SetLevel(4)
	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
	log.SetReportCaller(true)
}

func initDB() {
	setting.EtcdSetting = &setting.Etcd{
		Endpoints: []string{EtcdEndpoint},
	}
	db.InitKV()
}