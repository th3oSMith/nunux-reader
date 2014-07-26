package main

import (
	"encoding/base64"
	"github.com/zenazn/goji/web"
	"net/http"
	"strings"
)

// Création des utilisateurs
var users map[string]string

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

		if err != nil || isUser(string(credentials)) != true {
			pleaseAuth(w)
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

func initUsers() {
	users = make(map[string]string)
	users["admin"] = "admin"
	users["airremi"] = "abc"
}

func isUser(credentials string) bool {

	for user, password := range users {
		if credentials == user+":"+password {
			return true
		}
	}
	return false
}
