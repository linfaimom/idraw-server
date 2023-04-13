package db

import (
	"log"
	"os"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var dbInstance *gorm.DB

func init() {
	log.Println("validating db connections env injections")
	dbPath := os.Getenv("SQLITE_DB_PATH")
	if dbPath == "" {
		log.Fatalln("lack env SQLITE_DB_PATH")
	}
	log.Println("validation done")
	var err error
	dbInstance, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalln("open sqlite failed")
	}
}
