package storage

import (
	"database/sql"
	"encoding/json"
	_ "github.com/go-sql-driver/mysql"
	"github.com/th3osmith/rss"
	"io/ioutil"
	"log"
	"os"
)

var db *sql.DB

func Init(sqlDB *sql.DB) {

	db = sqlDB
	log.Println("Module de stockage chargé")

	err := LoadUsers()
	if err != nil {
		log.Println("Impossible de charger le module utilisateurs")
		log.Fatal(err)
	}
	log.Println("Module Utilisateur chargé")

	err = LoadFeeds()
	if err != nil {
		log.Println("Impossible de charger le module Flux")
		log.Fatal(err)
	}
	log.Println("Module Flux chargé")

	err = LoadTimelines()
	if err != nil {
		log.Println("Impossible de charger le module Timelines")
		log.Fatal(err)
	}
	log.Println("Module Timelines chargé")

	known, err := RecoverKnown()
	if err != nil {
		log.Println("Impossible de récupérer l'état du parseur")
		log.Fatal(err)
	}
	log.Println("État du parseur chargé")

	if len(known) > 0 {
		rss.Restore(known)
	}

}

func RecoverKnown() (known map[string]struct{}, err error) {

	if _, err := os.Stat("known.db"); err != nil {
		log.Println("Pas de fichier db trouvé")
		return known, nil
	}

	jsonData, err := ioutil.ReadFile("known.db")
	if err != nil {
		return known, err
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
