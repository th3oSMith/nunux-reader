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

type Context struct {
	User      User
	Feeds     map[int64]*rss.Feed
	Timelines map[int64]*Timeline
	Archive   *Timeline
}

var db *sql.DB

// Paramètres
var MaxArticles int
var UpdateTime, DeleteTime int

func Init(sqlDB *sql.DB, p Parameters) {

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

	// Initialisation des paramètres
	MaxArticles = p.MaxArticles
	UpdateTime = p.UpdateTime
	DeleteTime = p.DeleteTime

	// Initialisation des map pour les utilisateurs
	UserFeeds = make(map[int64]map[int64]*rss.Feed)
	UserTimelines = make(map[int64]map[int64]*Timeline)
	CurrentUsers = make(map[string]User)

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

type DatabaseParameters struct {
	Host     string `json:"host"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
}

type Parameters struct {
	Database    DatabaseParameters `json:"database"`
	UpdateTime  int                `json:"updateTime"`
	MaxArticles int                `json:"maxArticles"`
	DeleteTime  int                `json:"deleteTime"`
}
