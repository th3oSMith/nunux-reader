package main

import (
	"encoding/json"
	"github.com/th3osmith/nunux-reader/storage"
	"github.com/th3osmith/rss"
	"github.com/zenazn/goji/web"
	"io"
	"log"
	"net/http"
	"time"
)

type Feed struct {
	Title        string    `json:"title"`
	UpdateUrl    string    `json:"xmlurl"`
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

type TimelineData struct {
	Next     string     `json:"next"`
	Articles []rss.Item `json:"articles"`
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

	subs := storage.Feeds

	b, err := json.MarshalIndent(subs, "", "    ")

	if err != nil {
		log.Fatal(err)
	}

	io.WriteString(w, string(b))

}

func TimelinePage(w http.ResponseWriter, r *http.Request) {

	timelines := storage.Timelines

	b, err := json.MarshalIndent(timelines, "", "    ")

	if err != nil {
		log.Fatal(err)
	}

	io.WriteString(w, string(b))

}

func TimelineStatus(c web.C, w http.ResponseWriter, r *http.Request) {

	timelineName := c.URLParams["name"]
	log.Println(timelineName)

	timeline, err := storage.GetTimeline(timelineName)
	if err != nil {
		log.Fatal(err)
	}

	b, err := json.MarshalIndent(timeline, "", "    ")

	if err != nil {
		log.Fatal(err)
	}

	io.WriteString(w, string(b))
}

func getTimeline(c web.C, w http.ResponseWriter, r *http.Request) {

	timelineName := c.URLParams["name"]
	log.Println(timelineName)

	timeline, err := storage.GetTimeline(timelineName)
	if err != nil {
		log.Fatal(err)
	}

	articles, err := storage.GetFeedArticles(timeline.Feed.Id)
	if err != nil {
		log.Fatal(err)
	}

	data := TimelineData{"next?", articles}

	b, err := json.MarshalIndent(data, "", "    ")

	if err != nil {
		log.Fatal(err)
	}

	io.WriteString(w, string(b))

}
