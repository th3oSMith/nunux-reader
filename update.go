package main

import (
	"github.com/th3osmith/nunux-reader/storage"
	"github.com/th3osmith/rss"
	"log"
	"time"
)

var up *updater

type updater struct {
	quit chan struct{}
}

func InitUpdater() {
	up = NewUpdater()
	go up.Run()
}

func NewUpdater() *updater {
	out := new(updater)
	out.quit = make(chan struct{})
	return out
}

func (u *updater) Run() {
	ticker := time.NewTicker(1 * time.Minute)
	Update()

	for {
		select {
		case <-ticker.C:
			log.Println("Lancement de la mise à jour")
			Update()
		case <-u.quit:
			ticker.Stop()
			return
		}
	}
}

func Update() (err error) {

	for _, feed := range storage.Feeds {

		log.Println("Mise à jour du Flux ", feed.Title)

		articles, err := feed.GetNew()

		if err != nil {
			return err
		}
		log.Println("Récupération de ", len(articles), "articles")

		err = storage.SaveArticles(articles, feed.Id)
		if err != nil {
			return err
		}
	}

	// Sauvegarde de l'état du parseur RSS
	log.Println("Sauvegarde de l'état du parseur")
	storage.SaveKnown(rss.GetState())

	return err

}

func (u *updater) Quit() {
	close(u.quit)
}
