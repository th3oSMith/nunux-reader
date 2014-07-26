package main

import (
	"encoding/json"
	"github.com/zenazn/goji/web"
	"io"
	"log"
	"net/http"
	"time"
)

type Feed struct {
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

type Timeline struct {
	Timeline string `json:"timeline"`
	Title    string `json:"title"`
	Size     int    `json:"size"`
	Feed     Feed   `json:"feed"`
}

type TimelineData struct {
	Next     string    `json:"next"`
	Articles []Article `json:"articles"`
}

type Article struct {
	Author      string    `json:"author"`
	Date        time.Time `json:"date"`
	Description string    `json:"description"`
	Guid        string    `json:"guid"`
	Id          string    `json:"id"`
	Link        string    `json:"link"`
	Origlink    string    `json:"origlink"`
	Pubdate     time.Time `json:"pubdate"`
	Summary     string    `json:"summary"`
	Title       string    `json:"title"`
	Fid         string    `json:"fid"`
}

func SubscriptionPage(w http.ResponseWriter, r *http.Request) {

	subs := []Feed{
		{"Titre", "wml", "status", time.Now(), time.Now(), "etag", time.Now(), 0, "descr", "link", "hub", "iddd"},
		{"Titre2", "wml", "status", time.Now(), time.Now(), "etag", time.Now(), 0, "descr", "link", "hub", "iddd"},
	}

	b, err := json.MarshalIndent(subs, "", "    ")

	if err != nil {
		log.Fatal(err)
	}

	io.WriteString(w, string(b))

}

func TimelinePage(w http.ResponseWriter, r *http.Request) {

	timelines := []Timeline{
		{"global", "Titre", 23, Feed{}},
		{"archive", "Saved Items", 42, Feed{}},
	}

	b, err := json.MarshalIndent(timelines, "", "    ")

	if err != nil {
		log.Fatal(err)
	}

	io.WriteString(w, string(b))

}

func TimelineStatus(c web.C, w http.ResponseWriter, r *http.Request) {

	timelineName := c.URLParams["name"]
	log.Println(timelineName)

	timeline := Timeline{"global", "Titre", 23, Feed{}}

	b, err := json.MarshalIndent(timeline, "", "    ")

	if err != nil {
		log.Fatal(err)
	}

	io.WriteString(w, string(b))
}

func getTimeline(c web.C, w http.ResponseWriter, r *http.Request) {

	timelineName := c.URLParams["name"]
	log.Println(timelineName)

	data := TimelineData{"next?", []Article{
		{"auithor", time.Now(), "desc", "guid", "id", "orig", "pubd", time.Now(), "till", "fid", "ttt"},
		{"auithora", time.Now(), "desc", "guid", "id", "orig", "pubd", time.Now(), "till", "fid", "ttt"},
	},
	}

	b, err := json.MarshalIndent(data, "", "    ")

	if err != nil {
		log.Fatal(err)
	}

	io.WriteString(w, string(b))

}
