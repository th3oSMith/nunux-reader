package main

import (
	"github.com/th3osmith/nunux-reader/storage"
	"github.com/th3osmith/rss"
	"log"
)

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
