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

	userCredentials := strings.Split(credentials, ":")

	if len(userCredentials) != 2 {
		return false
	}

	for _, user := range storage.Users {
		if userCredentials[0] == user.Username && storage.Sha256Sum(userCredentials[1]) == user.Password {
			userWithPwd := user
			userWithPwd.Password = userCredentials[1]
			storage.CurrentUsers[auth] = userWithPwd
			storage.InitUser(userWithPwd)
			storage.UpdateUser(user.Id) // Magic
			return true
		}
	}

	return false
}
