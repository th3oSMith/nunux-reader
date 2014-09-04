package main

import (
	"database/sql"
	"encoding/json"
	_ "github.com/go-sql-driver/mysql"
	"github.com/th3osmith/nunux-reader/storage"
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var db *sql.DB

func main() {

	log.Println("Lancement de Nunux-Reader")
	var err error

	log.Println("Chargement des paramètres")
	var p storage.Parameters
	file, err := os.Open("parameters.json")
	if err != nil {
		log.Fatal(err)
	}

	err = json.NewDecoder(file).Decode(&p)
	if err != nil {
		log.Fatal(err)
	}

	// Ouverture de la connexion à la base SQL
	log.Println("Ouverture de la connexion MySQL")
	connexionInfos := p.Database.User + ":" + p.Database.Password + "@tcp(" + p.Database.Host + ":3306)/" + p.Database.Database + "?parseTime=1"
	db, err = sql.Open("mysql", connexionInfos)

	if err != nil {
		log.Fatal(err)
	}

	// Test de la connection à la base de données
	err = db.Ping()
	if err != nil {
		log.Println("Impossible de se connecter à la base de données")
		log.Fatal(err)
	}

	defer db.Close()

	// Initialisation des modules
	log.Println("---Initialisation des modules---")
	storage.Init(db, p)
	InitUpdater()
	InitDeleter()

	// Définition des routes

	// Accueil du site
	goji.Get("/", Root)
	goji.Use(SuperSecure)

	// Page nécessitant une authentification
	api := web.New()
	goji.Handle("/api/*", api)
	api.Use(SuperSecure)

	api.Get("/api/subscription", SubscriptionPage)
	api.Get("/api/timeline", TimelinePage)
	api.Get("/api/timeline/", http.RedirectHandler("/api/timeline", 301))
	api.Get("/api/timeline/:name/status", TimelineStatus)
	api.Get("/api/timeline/:name", getTimeline)

	api.Post("/api/subscription", addSubscription)
	api.Get("/api/subscription/export", exportOPML)
	api.Post("/api/subscriptionOPML", addOPML)
	api.Delete("/api/subscription/:id", removeSubscription)

	api.Put("/api/timeline/archive/:id", saveArticle)

	api.Delete("/api/timeline/:timeline/:id", removeArticle)
	api.Put("/api/timeline/:timeline/:id", recoverArticle)
	api.Delete("/api/timeline/:timeline", removeTimelineArticles)

	// REST User

	admin := web.New()
	goji.Handle("/admin/*", admin)
	admin.Use(SuperSecure)
	admin.Use(Admin)

	admin.Get("/admin/user", getUsers)
	admin.Get("/admin/user/current", getCurrentUser)
	admin.Get("/admin/user/:userId", getUser)
	admin.Post("/admin/user", createUser)
	admin.Put("/admin/user/:userId", updateUser)
	admin.Delete("/admin/user/:userId", deleteUser)

	// Application Angular
	// On le met en dernier pour ne pas pourrir toutes les routes
	goji.Get("/*", http.FileServer(http.Dir("public")))

	goji.Serve()

}

func Root(w http.ResponseWriter, r *http.Request) {

	// Rechargement des timelines
	storage.LoadTimelines()

	// Chargement de la page
	body, err := ioutil.ReadFile("views/index.html")

	if err != nil {
		log.Println("Impossible de lire le contenu statique du site")
		log.Fatal(err)
	}

	io.WriteString(w, string(body))

}

func AdminRoot(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Il n'y a rien à voir ici")
}
