package storage

import (
	"encoding/xml"
	"github.com/th3osmith/rss"
	"log"
)

func AddOPML(opml []byte, c Context) (feeds []rss.Feed, err error) {

	log.Println("Récupération de l'OPML")

	v := Opml{}
	err = xml.Unmarshal(opml, &v)

	if err != nil {
		return
	}

	for _, x := range v.Body.Outline.Outline {
		// Création du Flux
		feed, err := CreateFeed(x.XmlUrl, c, false, rss.Credentials{})
		if err != nil {
			log.Fatal(err)
		}

		feeds = append(feeds, *feed)

		// Création de la Timeline
		err = CreateTimeline(feed.Title, feed, c)
		if err != nil {
			log.Fatal(err)
		}
	}

	return
}

func ExportOPML() (opml *Opml) {

	var outlines []Outline

	for _, timeline := range Timelines {
		outlines = append(outlines, Outline{Type: "rss", Text: timeline.Timeline, Title: timeline.Timeline, Description: timeline.Feed.Description, XmlUrl: timeline.Feed.UpdateURL})
	}

	outline := Outline{Outline: outlines}

	head := Head{Title: "Nunux Export"}
	body := Body{Outline: outline}

	opml = &Opml{Version: "1.0", Head: head, Body: body}

	return opml
}

type Opml struct {
	Head    Head   `xml:"head"`
	Body    Body   `xml:"body"`
	Version string `xml:"version,attr"`
}

type Head struct {
	Title string `xml:"title"`
}

type Body struct {
	Outline Outline `xml:"outline"`
}

type Outline struct {
	Text        string    `xml:"text,attr"`
	Type        string    `xml:"type,attr"`
	Title       string    `xml:"title,attr"`
	Description string    `xml:"description,attr"`
	XmlUrl      string    `xml:"xmlUrl,attr"`
	Outline     []Outline `xml:"outline"`
}
