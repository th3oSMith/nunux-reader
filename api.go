package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"
)

func SubscriptionPage(w http.ResponseWriter, r *http.Request) {

	type Subscription struct {
		Title        string    `json:"title"`
		Xmlurl       string    `json:"xmlurl"`
		Status       string    `json:"status"`
		LastModified time.Time `json:"lastModified"`
		Expires      time.Time `json:"expires"`
		Etag         string    `json:"etag"`
		UpdateDate   time.Time `json:"updateDate"`
		ErrCount     int       `json:"errCount"`
		Description  string    `json:"description"`
		Link         string    `json:"link"`
		Hub          string    `json:"hub"`
		Id           string    `json:"id"`
	}

	subs := []Subscription{
		{"Titre", "wml", "status", time.Now(), time.Now(), "etag", time.Now(), 0, "descr", "link", "hub", "iddd"},
		{"Titre2", "wml", "status", time.Now(), time.Now(), "etag", time.Now(), 0, "descr", "link", "hub", "iddd"},
	}

	b, err := json.MarshalIndent(subs, "", "    ")

	if err != nil {
		log.Fatal(err)
	}

	io.WriteString(w, string(b))

}
