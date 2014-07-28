package main

import (
	"encoding/json"
	"github.com/th3osmith/nunux-reader/storage"
	"github.com/th3osmith/rss"
	"github.com/zenazn/goji/web"
	"io"
	"log"
	"net/http"
	"strconv"
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

	var subs []*rss.Feed

	for _, f := range storage.Feeds {
		subs = append(subs, f)
	}

	b, err := json.MarshalIndent(subs, "", "    ")

	if err != nil {
		log.Fatal(err)
	}

	io.WriteString(w, string(b))

}

func TimelinePage(w http.ResponseWriter, r *http.Request) {

	var timelines []storage.Timeline

	// On ajoute les timelines spéciales
	size, err := storage.GetGlobalArticlesSize()
	if err != nil {
		log.Fatal(err)
	}
	global := storage.Timeline{"global", "All items", size, rss.Feed{}, -1}

	timelines = append(timelines, global)

	// On traite les autres
	for _, t := range storage.Timelines {
		timelines = append(timelines, t)
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

	b := getTimelineStatus(timelineName)

	io.WriteString(w, b)
}

func getTimelineStatus(timelineName string) (status string) {

	var err error
	var timeline storage.Timeline

	// Traitement des cas particuliers
	if timelineName == "global" {

		size, err := storage.GetGlobalArticlesSize()
		if err != nil {
			log.Fatal(err)
		}

		timeline = storage.Timeline{"global", "All items", size, rss.Feed{}, -1}

	} else {
		timeline, err = storage.GetTimeline(timelineName)
	}

	if err != nil {
		log.Fatal(err)
	}

	b, err := json.MarshalIndent(timeline, "", "    ")

	if err != nil {
		log.Fatal(err)
	}

	return string(b)

}

func getTimeline(c web.C, w http.ResponseWriter, r *http.Request) {

	timelineName := c.URLParams["name"]
	log.Println(timelineName)

	var articles []rss.Item
	var err error

	// Gestion des cas particuliers
	if timelineName == "global" {
		articles, err = storage.GetGlobalArticles()
		if err != nil {
			log.Fatal(err)
		}

	} else {

		timelineId, _ := strconv.Atoi(timelineName)

		articles, err = storage.GetTimelineArticles(int64(timelineId))
		if err != nil {
			log.Fatal(err)
		}
	}
	data := TimelineData{"next?", articles}

	b, err := json.MarshalIndent(data, "", "    ")

	if err != nil {
		log.Fatal(err)
	}

	io.WriteString(w, string(b))

}

type receiver struct {
	Url string
}

func addSubscription(w http.ResponseWriter, r *http.Request) {

	// Récpération de l'url envoyée dans le corps en json
	var v receiver
	json.NewDecoder(r.Body).Decode(&v)

	// Création du Flux
	feed, err := storage.CreateFeed(v.Url)
	if err != nil {
		log.Fatal(err)
	}

	// Création de la Timeline
	err = storage.CreateTimeline(feed.Title, feed)
	if err != nil {
		log.Fatal(err)
	}

	// Création de la sortie
	out, err := json.MarshalIndent(feed, "", "    ")
	if err != nil {
		log.Fatal(err)
	}

	io.WriteString(w, string(out))

}

func removeSubscription(c web.C, w http.ResponseWriter, r *http.Request) {

	id := c.URLParams["id"]

	log.Println("Suppression du flux", id)

	idInt, _ := strconv.Atoi(id)

	err := storage.RemoveFeed(int64(idInt))
	if err != nil {
		log.Fatal(err)
	}

	storage.LoadFeeds()

}

func removeArticle(c web.C, w http.ResponseWriter, r *http.Request) {

	id := c.URLParams["id"]
	timeline := c.URLParams["timeline"]
	log.Println("Suppression de l'article ", id)
	idInt, _ := strconv.Atoi(id)

	err := storage.RemoveArticle(int64(idInt), timeline)
	if err != nil {
		log.Fatal(err)
	}

	b := getTimelineStatus(timeline)

	io.WriteString(w, b)

}
