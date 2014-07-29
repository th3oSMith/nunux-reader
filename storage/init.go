package storage

import (
	"database/sql"
	"encoding/json"
	_ "github.com/go-sql-driver/mysql"
	"github.com/th3osmith/rss"
	"io/ioutil"
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

	err = LoadTimelines()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Module Timelines chargé")

	known, err := RecoverKnown()
	if err != nil {
		log.Fatal(err)
	}

	rss.Restore(known)

}

func RecoverKnown() (known map[string]struct{}, err error) {

	jsonData, err := ioutil.ReadFile("known.db")
	if err != nil {
		return
	}

	json.Unmarshal(jsonData, &known)

	return

}

func SaveKnown(known map[string]struct{}) (err error) {

	data, err := json.Marshal(known)
	if err != nil {
		return
	}

	ioutil.WriteFile("known.db", data, 0600)

	return
}
