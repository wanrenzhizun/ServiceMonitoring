package dbUtil

import (
	"fmt"
	"github.com/asdine/storm/v3"
	"os"
	"project/server/config"
)

type ExecuteFunc func(db *storm.DB) error

var db *storm.DB
var logDb *storm.DB
func init() {
	OpenLDb()
	OpenLogDb()
}



func OpenLogDb()  {
	var err error
	logDb, err = storm.Open(config.Cfg.Section("common").Key("baseDir").String() + "log/log.db" ,storm.Batch())
	if err != nil{
		fmt.Println(err.Error())
	}
}
func OpenLDb()  {
	var err error
	db, err = storm.Open(config.Cfg.Section("common").Key("baseDir").String() + "data/source.db",storm.Batch())
	if err != nil{
		fmt.Println(err.Error())
	}
}

func CloseDb()  {
	db.Close()
}

func CloseLogDb()  {
	logDb.Close()
}

func Execute(exec ExecuteFunc) error {
	err := exec(db)
	return err
}

func ExecuteLog(exec ExecuteFunc) error {
	err := exec(logDb)
	return err
}

func DeleteLogFile()  {
	CloseLogDb()
	err := os.Remove("log/log.db")
	if err != nil {
		fmt.Println(err)
	}
	OpenLogDb()

}

