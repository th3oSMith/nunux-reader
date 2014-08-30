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
)

type TimelineData struct {
	Next     int64      `json:"next"`
	Articles []rss.Item `json:"articles"`
}

type OPMLError struct {
	ErrorMsg string `json:"error"`
}

func getContext(r *http.Request) (c storage.Context) {

	c.User = storage.CurrentUsers[r.Header.Get("Authorization")]
	c.Feeds = storage.UserFeeds[c.User.Id]
	c.Timelines = storage.UserTimelines[c.User.Id]
	c.Archive = storage.Archives[c.User.SavedTimelineId]

	return

}

// Récupération des abonnements de l'utilisateur
func SubscriptionPage(w http.ResponseWriter, r *http.Request) {

	var subs []*rss.Feed
	c := getContext(r)

	for _, f := range c.Feeds {
		subs = append(subs, f)
	}

	b, err := json.MarshalIndent(subs, "", "    ")

	if err != nil {
		log.Println(err)
		http.Error(w, "Impossible de récupérer les abonnements de l'utilisateur", 500)
		return
	}

	io.WriteString(w, string(b))

}

// Récupération des timelines de l'utilisateur
func TimelinePage(w http.ResponseWriter, r *http.Request) {

	var timelines []storage.Timeline
	c := getContext(r)

	// On ajoute les timelines spéciales
	size, err := storage.GetGlobalArticlesSize(c)
	if err != nil {
		log.Println("Impossible de récupérer la taille de la timeline globale")
		log.Println(err)
	}
	global := storage.Timeline{"global", "All items", size, rss.Feed{}, -1}

	timelines = append(timelines, global)
	timelines = append(timelines, *c.Archive)

	// On traite les autres
	for _, t := range c.Timelines {
		timelines = append(timelines, *t)
	}

	b, err := json.MarshalIndent(timelines, "", "    ")

	if err != nil {
		log.Println(err)
		http.Error(w, "Impossible de récupérer les timelines", 500)
		return
	}

	io.WriteString(w, string(b))

}

// Récupération du statut d'une timeline
func TimelineStatus(c web.C, w http.ResponseWriter, r *http.Request) {

	context := getContext(r)
	timelineName := c.URLParams["name"]

	b := getTimelineStatus(timelineName, context)

	io.WriteString(w, b)
}

func getTimelineStatus(timelineName string, c storage.Context) (status string) {

	var err error
	var timeline storage.Timeline

	// Traitement des cas particuliers
	if timelineName == "global" {

		size, err := storage.GetGlobalArticlesSize(c)
		if err != nil {
			log.Println("Impossible de récupérer la taille de la timeline globale")
			log.Println(err)
		}

		timeline = storage.Timeline{"global", "All items", size, rss.Feed{}, -1}

	} else {
		timeline, err = storage.GetTimeline(timelineName, c)
	}

	if err != nil {
		log.Println("Impossible de récupérer la timeline")
		log.Println(err)
	}

	b, err := json.MarshalIndent(timeline, "", "    ")

	if err != nil {
		log.Println(err)
		log.Println("Impossible de récupérer la timeline")
	}

	return string(b)

}

func getTimeline(c web.C, w http.ResponseWriter, r *http.Request) {

	context := getContext(r)

	timelineName := c.URLParams["name"]

	var err error
	var nextId int

	if len(r.URL.Query()["next"]) > 0 {
		nextId, _ = strconv.Atoi(r.URL.Query()["next"][0])
	}

	var articles []rss.Item

	// Gestion des cas particuliers
	if timelineName == "global" {
		articles, err = storage.GetGlobalArticles(int64(nextId), context)
		if err != nil {
			log.Println(err)
			http.Error(w, "Impossible de récupérer la timeline globale", 500)
			return

		}
	} else if timelineName == "archive" {
		timelineId := context.Archive.Id
		articles, err = storage.GetTimelineArticles(int64(timelineId), int64(nextId))
		if err != nil {
			log.Println(err)
			http.Error(w, "Impossible de récupérer la timeline archive", 500)
			return
		}

	} else {

		timelineId, _ := strconv.Atoi(timelineName)

		articles, err = storage.GetTimelineArticles(int64(timelineId), int64(nextId))
		if err != nil {
			log.Println(err)
			http.Error(w, "Impossible de récupérer la timeline", 500)
			return
		}
	}

	var next int64
	if len(articles) > 0 && len(articles) == storage.MaxArticles {
		next = articles[len(articles)-1].Id
	}

	data := TimelineData{next, articles}

	if len(articles) == 0 {
		data.Articles = []rss.Item{}
	}

	b, err := json.MarshalIndent(data, "", "    ")

	if err != nil {
		log.Println(err)
	}

	io.WriteString(w, string(b))

}

type receiver struct {
	Url string
}

func addSubscription(w http.ResponseWriter, r *http.Request) {

	c := getContext(r)

	// Récpération de l'url envoyée dans le corps en json
	var v receiver
	json.NewDecoder(r.Body).Decode(&v)

	// Création du Flux
	feed, err := storage.CreateFeed(v.Url, c)
	if err != nil {
		log.Println(err)
		http.Error(w, "Impossible de créer le flux", 500)
		return
	}

	// Création de la Timeline
	err = storage.CreateTimeline(feed.Title, feed, c)
	if err != nil {
		log.Println(err)
		http.Error(w, "Impossible de créer la timeline", 500)
		return
	}

	// Création de la sortie
	out, err := json.MarshalIndent(feed, "", "    ")
	if err != nil {
		log.Println(err)
	}

	io.WriteString(w, string(out))

}

func addOPML(w http.ResponseWriter, r *http.Request) {

	context := getContext(r)

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

	feeds, err := storage.AddOPML(opml, context)
	if err != nil {
		uploadError(w, err)
		return
	}

	// Création de la sortie
	out, err := json.MarshalIndent(feeds, "", "    ")
	if err != nil {
		log.Println(err)
		http.Error(w, "Erreur lors de l'importation", 500)
		return
	}

	io.WriteString(w, string(out))

}

func exportOPML(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Disposition", "attachment; filename=nunux.opml")
	w.Header().Set("Content-Type", "text/x-opml+xml")

	opml := storage.ExportOPML()

	out, err := xml.MarshalIndent(opml, "", "    ")
	if err != nil {
		log.Println(err)
	}

	io.WriteString(w, `<?xml version="1.0" encoding="UTF-8"?>`)
	io.WriteString(w, string(out))
}

func uploadError(w http.ResponseWriter, errorMsg error) {

	output := OPMLError{ErrorMsg: errorMsg.Error()}
	out, err := json.MarshalIndent(output, "", "    ")

	http.Error(w, err.Error(), 500)

	if err != nil {
		log.Println(err)
	}

	io.WriteString(w, string(out))

}

func removeSubscription(c web.C, w http.ResponseWriter, r *http.Request) {

	context := getContext(r)
	id := c.URLParams["id"]

	idInt, _ := strconv.Atoi(id)

	err := storage.RemoveTimeline(int64(idInt), context)
	if err != nil {
		log.Println(err)
		http.Error(w, "Impossible de supprimer l'abonnement", 500)
	}

	storage.LoadFeeds()

}

func removeArticle(c web.C, w http.ResponseWriter, r *http.Request) {

	context := getContext(r)

	id := c.URLParams["id"]
	timeline := c.URLParams["timeline"]
	idInt, _ := strconv.Atoi(id)

	err := storage.SoftRemoveArticle(int64(idInt), timeline, context)
	if err != nil {
		log.Println(err)
		http.Error(w, "Impossible de supprimer l'article", 500)
		return
	}

	b := getTimelineStatus(timeline, context)

	io.WriteString(w, b)

}

func recoverArticle(c web.C, w http.ResponseWriter, r *http.Request) {

	context := getContext(r)

	id := c.URLParams["id"]
	timeline := c.URLParams["timeline"]
	idInt, _ := strconv.Atoi(id)

	err := storage.RecoverArticle(int64(idInt), timeline, context)
	if err != nil {
		log.Println(err)
		http.Error(w, "Impossible de récupérer l'article", 500)
		return
	}

	b := getTimelineStatus(timeline, context)

	io.WriteString(w, b)

}

func saveArticle(c web.C, w http.ResponseWriter, r *http.Request) {

	context := getContext(r)

	id := c.URLParams["id"]
	idInt, _ := strconv.Atoi(id)

	err := storage.SaveArticle(int64(idInt), context)
	if err != nil {
		log.Println(err)
		http.Error(w, "Impossible de sauvegarder l'article", 500)
		return
	}

	b := getTimelineStatus("archive", context)

	io.WriteString(w, b)

}

func removeTimelineArticles(c web.C, w http.ResponseWriter, r *http.Request) {

	context := getContext(r)
	timeline := c.URLParams["timeline"]

	err := storage.RemoveTimelineArticles(timeline)
	if err != nil {
		log.Println(err)
		http.Error(w, "Impossible de supprimer les articles de la timeline", 500)
		return
	}

	b := getTimelineStatus(timeline, context)

	io.WriteString(w, b)

}
