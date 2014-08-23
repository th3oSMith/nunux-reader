package storage

import (
	"database/sql"
	"github.com/th3osmith/rss"
	"log"
	"strconv"
)

var Feeds map[int64]*rss.Feed

func CreateFeed(url string) (feed *rss.Feed, err error) {

	feed, err = rss.Fetch(url)

	if err != nil {
		log.Fatal(err)
	}

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
	Feeds[feed.Id] = feed
	Feeds[feed.Id].Items = []*rss.Item{}

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

	Feeds = make(map[int64]*rss.Feed)

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
		Feeds[feed.Id] = &feed
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

func SaveArticle(id int64) (err error) {

	// On enregistre l'article dans la timeline des sauvegardes de l'utilsateur
	stmt2, err := db.Prepare("INSERT INTO article_timelines(article_id, timeline_id) VALUES(?, ?)")

	timelineId := CurrentUser.SavedTimelineId
	_, err = stmt2.Exec(id, timelineId)
	if err != nil {
		return err
	}
	log.Println("Sauvegarde d'un article")
	Archive.Size++

	return nil

}

func SaveArticles(articles []*rss.Item, feedId int64) (err error) {

	// Récupération des timelinse qui possèdent ce flux

	var timelinesId []int64
	var id int64

	rows, err := db.Query("select id FROM timeline WHERE feed_id = ?", feedId)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&id)
		if err != nil {
			return err
		}
		timelinesId = append(timelinesId, id)
	}
	err = rows.Err()
	if err != nil {
		return err
	}

	stmt, err := db.Prepare("INSERT INTO article(date, description, link, pubdate, title, feed_id) VALUES(NOW(), ?, ?, ?, ?, ?)")
	stmt2, err := db.Prepare("INSERT INTO article_timelines(article_id, timeline_id) VALUES(?, ?)")

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

		// Insertion des articles dans les timelines
		for _, timelineId := range timelinesId {

			log.Println(lastId, timelineId)
			_, err = stmt2.Exec(lastId, timelineId)
			if err != nil {
				return err
			}
			log.Println("Insertion d'une Référence")

		}

	}

	// On met à jour les timelines en mémoire
	LoadTimelines()

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

func RemoveArticle(id int64, timelineName string) (err error) {

	sqlQ := "DELETE FROM article_timelines WHERE article_id = ? AND ("
	var args []interface{}

	args = append(args, id)

	if timelineName == "global" {
		for _, timeline := range Timelines {
			sqlQ += "timeline_id = ? OR "
			args = append(args, timeline.Id)
		}
		sqlQ = sqlQ[:len(sqlQ)-3] + ")"

	} else if timelineName == "archive" {

		timeId := CurrentUser.SavedTimelineId
		sqlQ += "timeline_id = ?) "
		args = append(args, timeId)
		Archive.Size--

	} else {
		tmp, _ := strconv.Atoi(timelineName)
		timeId := int64(tmp)
		sqlQ += "timeline_id = ?) "
		args = append(args, timeId)
		Timelines[timeId].Size--

	}

	stmt, err := db.Prepare(sqlQ)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(args...)
	if err != nil {
		return err
	}

	var number int64

	// Suppression de l'article si plus personne n'en a besoin
	err = db.QueryRow("select COUNT(*) number FROM article_timelines WHERE article_id = ?", id).Scan(&number)

	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if err == sql.ErrNoRows {
		return nil
	}

	if number == 0 {
		stmt, err = db.Prepare("DELETE FROM article WHERE id = ?;")

		if err != nil {
			return err
		}

		_, err = stmt.Exec(id)
		if err != nil {
			return err
		}
	}

	log.Printf("Suppression de l'article ID = %d", id)

	return nil

}

func RemoveTimelineArticles(timelineName string) (err error) {

	sql := "DELETE FROM article_timelines WHERE ("
	var args []interface{}

	if timelineName == "global" {
		for _, timeline := range Timelines {
			sql += "timeline_id = ? OR "
			args = append(args, timeline.Id)
		}
		sql = sql[:len(sql)-3] + ")"

	} else {
		tmp, _ := strconv.Atoi(timelineName)
		timeId := int64(tmp)
		sql += "timeline_id = ?) "
		args = append(args, timeId)
		Timelines[timeId].Size = 0

	}

	stmt, err := db.Prepare(sql)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(args...)
	if err != nil {
		return err
	}

	return nil

}
