package main

import (
	"encoding/base64"
	"github.com/th3osmith/nunux-reader/storage"
	"github.com/zenazn/goji/web"
	"net/http"
	"strings"
)

// Authentification basique
func SuperSecure(c *web.C, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")

		// On regarde si des identifiants ont été entrés
		if !strings.HasPrefix(auth, "Basic ") {
			pleaseAuth(w)
			return
		}

		// On regarde si les identifiants sont enregistrés
		credentials, err := base64.StdEncoding.DecodeString(auth[6:])

		if err != nil || isUser(string(credentials), auth) != true {
			pleaseAuth(w)
			return
		}

		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func Admin(c *web.C, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")

		user := storage.CurrentUsers[auth]

		if user.Id != 1 {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Go away!\n"))
			return
		}

		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func pleaseAuth(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="Nunux-Reader"`)
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte("Go away!\n"))
}

func isUser(credentials string, auth string) bool {

	for _, user := range storage.Users {
		if credentials == user.Username+":"+user.Password {
			storage.CurrentUsers[auth] = user
			storage.InitUser(user)
			storage.UpdateUser(user.Id) // Magic
			return true
		}
	}

	return false
}
