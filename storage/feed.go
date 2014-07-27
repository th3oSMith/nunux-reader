package storage

import (
	"github.com/th3osmith/rss"
	"log"
)

var Feeds []*rss.Feed

func CreateFeed(url string) (err error) {

	feed, err := rss.Fetch(url)

	if err != nil {
		log.Fatal(err)
	}

	// Ajout à la liste des flux chargés
	Feeds = append(Feeds, feed)

	// Enregistrement des articles
	// @TODO

	// Enregistrement dans la base SQL
	stmt, err := db.Prepare("INSERT INTO feed(nickname, title, description, link, updateUrl, refresh, unread) VALUES(?, ?, ?, ?, ?, ?, ?)")

	if err != nil {
		return err
	}

	res, err := stmt.Exec(feed.Nickname, feed.Title, feed.Description, feed.Link, feed.UpdateURL, feed.Refresh, feed.Unread)
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

func LoadFeeds() (err error) {

	log.Println("Chargement des flux")

	var feed rss.Feed

	rows, err := db.Query("select * from feed ")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&feed.Id, &feed.Nickname, &feed.Title, &feed.Description, &feed.Link, &feed.UpdateURL, &feed.Refresh, &feed.Unread)
		if err != nil {
			return err
		}
		Feeds = append(Feeds, &feed)
		log.Println(feed)
	}
	err = rows.Err()
	if err != nil {
		return err
	}

	return nil
}
