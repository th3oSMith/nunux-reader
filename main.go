package main

import (
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

func main() {

	initUsers()

	log.Println("Lancement de Nunux-Reader")

	// Définition des routes

	// Accueil du site
	goji.Get("/", Root)

	// Page nécessitant une authentification
	api := web.New()
	goji.Handle("/api/*", api)
	api.Use(SuperSecure)

	api.Get("/api/subscription", SubscriptionPage)
	api.Get("/api/timeline", TimelinePage)
	api.Get("/api/timeline/", http.RedirectHandler("/api/timeline", 301))
	api.Get("/api/timeline/:name/status", TimelineStatus)

	// Application Angular
	// On le met en dernier pour ne pas pourrir toutes les routes
	goji.Get("/*", http.FileServer(http.Dir("public")))

	goji.Serve()

}

func Root(w http.ResponseWriter, r *http.Request) {

	// Chargement de la page
	body, err := ioutil.ReadFile("views/index.html")

	if err != nil {
		log.Fatal(err)
	}

	io.WriteString(w, string(body))

}

func AdminRoot(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Il n'y a rien à voir ici")
}
