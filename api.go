package main

import (
	"encoding/json"
	"encoding/xml"
	"github.com/th3osmith/nunux-reader/storage"
	"github.com/th3osmith/rss"
	"github.com/zenazn/goji/web"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

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

type OPMLError struct {
	ErrorMsg string `json:"error"`
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
	timelines = append(timelines, storage.Archive)

	// On traite les autres
	for _, t := range storage.Timelines {
		timelines = append(timelines, *t)
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
	} else if timelineName == "archive" {
		timelineId := storage.CurrentUser.SavedTimelineId
		articles, err = storage.GetTimelineArticles(int64(timelineId))
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

	if len(articles) == 0 {
		data.Articles = []rss.Item{}
		data.Next = ""
	}

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

func addOPML(w http.ResponseWriter, r *http.Request) {

	file, _, err := r.FormFile("opml")
	if err != nil {
		uploadError(w, err)
		return
	}

	opml, err := ioutil.ReadAll(file)
	if err != nil {
		uploadError(w, err)
		return
	}

	feeds, err := storage.AddOPML(opml)
	if err != nil {
		uploadError(w, err)
		return
	}

	// Création de la sortie
	out, err := json.MarshalIndent(feeds, "", "    ")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Tttttt")

	io.WriteString(w, string(out))

}

func exportOPML(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Disposition", "attachment; filename=nunux.opml")
	w.Header().Set("Content-Type", "text/x-opml+xml")

	opml := storage.ExportOPML()

	out, err := xml.MarshalIndent(opml, "", "    ")
	if err != nil {
		log.Fatal(err)
	}

	io.WriteString(w, `<?xml version="1.0" encoding="UTF-8"?>`)
	io.WriteString(w, string(out))
}

func uploadError(w http.ResponseWriter, errorMsg error) {

	output := OPMLError{ErrorMsg: errorMsg.Error()}
	out, err := json.MarshalIndent(output, "", "    ")

	log.Println("Error", err)

	if err != nil {
		log.Fatal(err)
	}

	io.WriteString(w, string(out))

}

func removeSubscription(c web.C, w http.ResponseWriter, r *http.Request) {

	id := c.URLParams["id"]

	log.Println("Suppression de la timeline", id)

	idInt, _ := strconv.Atoi(id)

	err := storage.RemoveTimeline(int64(idInt))
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

func saveArticle(c web.C, w http.ResponseWriter, r *http.Request) {

	id := c.URLParams["id"]
	log.Println("Sauvegarde de l'article ", id)
	idInt, _ := strconv.Atoi(id)

	err := storage.SaveArticle(int64(idInt))
	if err != nil {
		log.Fatal(err)
	}

	b := getTimelineStatus("archive")

	io.WriteString(w, b)

}

func removeTimelineArticles(c web.C, w http.ResponseWriter, r *http.Request) {

	timeline := c.URLParams["timeline"]
	log.Println("Suppression des articles de la timeline ", timeline)

	err := storage.RemoveTimelineArticles(timeline)
	if err != nil {
		log.Fatal(err)
	}

	b := getTimelineStatus(timeline)

	io.WriteString(w, b)

}
