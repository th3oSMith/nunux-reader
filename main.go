package main

import (
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
	"io"
	"log"
	"net/http"
)

func main() {

	log.Println("Lancement de Nunux-Reader")

	// Définition des routes

	// Accueil du site
	goji.Get("/", Root)

	// Page nécessitant une authentification
	admin := web.New()
	goji.Handle("/admin/*", admin)
	admin.Use(SuperSecure)

	admin.Get("/admin/", AdminRoot)

	// Application Angular
	// On le met en dernier pour ne pas pourrir toutes les routes
	goji.Get("/*", http.FileServer(http.Dir("public")))

	goji.Serve()

}

func Root(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Il n'y a rien à voir ici")
}

func AdminRoot(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Il n'y a rien à voir ici")
}
