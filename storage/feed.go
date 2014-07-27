package storage

import (
	"github.com/SlyMarbo/rss"
	"log"
)

var feeds []*rss.Feed

func CreateFeed(url string) (err error) {

	feed, err := rss.Fetch(url)

	if err != nil {
		log.Fatal(err)
	}

	// Ajout à la liste des flux chargés
	feeds = append(feeds, feed)

	// Enregistrement dans la base SQL
	// Enregistrement dans la base SQL
	stmt, err := db.Prepare("INSERT INTO feed(nickname, title, description, link, updateUrl, unread) VALUES(?, ?, ?, ?, ?, ?)")

	if err != nil {
		return err
	}

	res, err := stmt.Exec(feed.Nickname, feed.Title, feed.Description, feed.Link, feed.UpdateURL, feed.Unread)
	if err != nil {
		return err
	}
	lastId, err := res.LastInsertId()

	if err != nil {
		return err
	}
	rowCnt, err := res.RowsAffected()

	if err != nil {
		return err
	}
	log.Printf("Insertion d'un Flux ID = %d, affected = %d\n", lastId, rowCnt)

	return nil

	return nil

}
