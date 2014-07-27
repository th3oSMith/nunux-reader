package storage

import (
	"github.com/th3osmith/rss"
	"log"
)

var Feeds []*rss.Feed

func CreateFeed(url string) (feed *rss.Feed, err error) {

	feed, err = rss.Fetch(url)

	if err != nil {
		log.Fatal(err)
	}

	// Enregistrement des articles
	// @TODO

	// Enregistrement dans la base SQL
	stmt, err := db.Prepare("INSERT INTO feed(nickname, title, description, link, updateUrl, refresh, unread) VALUES(?, ?, ?, ?, ?, ?, ?)")

	if err != nil {
		return nil, err
	}

	res, err := stmt.Exec(feed.Nickname, feed.Title, feed.Description, feed.Link, feed.UpdateURL, feed.Refresh, feed.Unread)
	if err != nil {
		return nil, err
	}
	lastId, err := res.LastInsertId()

	// Assignation de l'id
	feed.Id = lastId

	// Ajout à la liste des flux chargés
	Feeds = append(Feeds, feed)

	if err != nil {
		return nil, err
	}
	rowCnt, err := res.RowsAffected()

	if err != nil {
		return nil, err
	}
	log.Printf("Insertion d'un Flux ID = %d, affected = %d\n", lastId, rowCnt)

	return feed, nil

}

func LoadFeeds() (err error) {

	Feeds = nil

	log.Println("Chargement des flux")

	rows, err := db.Query("select * from feed ")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var feed rss.Feed
		err := rows.Scan(&feed.Id, &feed.Nickname, &feed.Title, &feed.Description, &feed.Link, &feed.UpdateURL, &feed.Refresh, &feed.Unread)
		if err != nil {
			return err
		}
		Feeds = append(Feeds, &feed)
	}
	err = rows.Err()
	if err != nil {
		return err
	}

	log.Println(Feeds)

	return nil
}

func GetFeedArticles(feedId int64) (articles []rss.Item, err error) {

	var article rss.Item

	log.Println("Chargement des articles du Flux", feedId)

	rows, err := db.Query("select id, date, description, link, pubdate, title from article where feed_id = ?", feedId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&article.Id, &article.Date, &article.Content, &article.Link, &article.PubDate, &article.Title)
		if err != nil {
			return nil, err
		}
		articles = append(articles, article)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return articles, nil
}

func SaveArticles(articles []*rss.Item, feedId int64) (err error) {

	stmt, err := db.Prepare("INSERT INTO article(date, description, link, pubdate, title, feed_id) VALUES(NOW(), ?, ?, ?, ?, ?)")

	if err != nil {
		return err
	}

	for _, article := range articles {
		res, err := stmt.Exec(article.Content, article.Link, article.Date, article.Title, feedId)
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
		log.Printf("Insertion d'un Article ID = %d, affected = %d\n", lastId, rowCnt)
	}
	return nil
}

func CountFeedArticles(feedId int64) (count int, err error) {

	err = db.QueryRow("SELECT COUNT(*) FROM article WHERE feed_id = ?", feedId).Scan(&count)

	if err != nil {
		return 0, err
	}

	return

}

func RemoveFeed(feedId int64) (err error) {

	// Suppression des Articles
	stmt, err := db.Prepare("DELETE FROM article WHERE feed_id = ?;")

	if err != nil {
		return err
	}

	_, err = stmt.Exec(feedId)
	if err != nil {
		return err
	}

	// Suppression du flux
	stmt, err = db.Prepare("DELETE FROM feed WHERE id = ?;")

	if err != nil {
		return err
	}

	_, err = stmt.Exec(feedId)
	if err != nil {
		return err
	}

	log.Printf("Suppression du flux ID = %d", feedId)

	return nil

}
