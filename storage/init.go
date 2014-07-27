package storage

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

var db *sql.DB

func Init(sqlDB *sql.DB) {

	db = sqlDB
	log.Println("Module de stockage chargé")

	err := LoadUsers()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Module Utilisateur chargé")

	err = LoadFeeds()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Module Flux chargé")

}
