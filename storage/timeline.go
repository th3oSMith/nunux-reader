package storage

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/th3osmith/rss"
	"log"
)

type Timeline struct {
	Timeline string   `json:"timeline"`
	Title    string   `json:"title"`
	Size     int      `json:"size"`
	Feed     rss.Feed `json:"feed"`
	Id       int64    `json:"id"`
}

var Timelines []Timeline

func GetTimeline(name string) (t Timeline, err error) {

	log.Println("Récupération de la timeline")

	err = db.QueryRow("select t.timeline, t.title, (SELECT COUNT(*) FROM article WHERE feed_id = t.feed_id) size, f.id, f.nickname, f.title, f.description, f.link, f.updateUrl, f.refresh, f.unread from timeline as t LEFT JOIN feed as f ON f.id = t.feed_id where t.id = ?", name).Scan(&t.Timeline, &t.Title, &t.Size, &t.Feed.Id, &t.Feed.Nickname, &t.Feed.Title, &t.Feed.Description, &t.Feed.Link, &t.Feed.UpdateURL, &t.Feed.Refresh, &t.Feed.Unread)

	if err != nil && err != sql.ErrNoRows {
		return Timeline{}, err
	}

	if err == sql.ErrNoRows {
		return Timeline{}, nil
	}

	return

}

func LoadTimelines() (err error) {

	// Initialization de la map
	log.Println("Chargement des Timelines")

	rows, err := db.Query("select t.id, t.timeline, t.title, (SELECT COUNT(*) FROM article WHERE feed_id = t.feed_id) size ,f.id, f.nickname, f.title, f.description, f.link, f.updateUrl, f.refresh, f.unread from timeline as t LEFT JOIN feed as f ON f.id = t.feed_id")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var t Timeline
		err := rows.Scan(&t.Id, &t.Timeline, &t.Title, &t.Size, &t.Feed.Id, &t.Feed.Nickname, &t.Feed.Title, &t.Feed.Description, &t.Feed.Link, &t.Feed.UpdateURL, &t.Feed.Refresh, &t.Feed.Unread)
		if err != nil {
			return err
		}
		Timelines = append(Timelines, t)
	}
	err = rows.Err()
	if err != nil {
		return err
	}

	return nil

}
