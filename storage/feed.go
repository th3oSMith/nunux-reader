package storage

import (
	"database/sql"
	"github.com/th3osmith/rss"
	"log"
	"strconv"
)

var Feeds map[int64]*rss.Feed
var UserFeeds map[int64]map[int64]*rss.Feed

func CreateFeed(url string, c Context, insecure bool, credentials rss.Credentials) (feed *rss.Feed, err error) {

	var id int64

	// On regarde si le flux existe déjà
	err = db.QueryRow("SELECT id FROM feed WHERE updateUrl = ?", url).Scan(&id)

	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	if err == sql.ErrNoRows {

		feed, err = rss.Fetch(url, insecure, credentials)

		if err != nil {
			return nil, err
		}

		// Enregistrement dans la base SQL
		stmt, err := db.Prepare("INSERT INTO feed(nickname, title, description, link, updateUrl, refresh, unread, insecure, username, password) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")

		if err != nil {
			return nil, err
		}

		res, err := stmt.Exec(feed.Nickname, feed.Title, feed.Description, feed.Link, feed.UpdateURL, feed.Refresh, feed.Unread, feed.Insecure, feed.Credentials.Username, feed.Credentials.Password)
		if err != nil {
			return nil, err
		}
		lastId, err := res.LastInsertId()

		// Assignation de l'id
		feed.Id = lastId

		// Ajout à la liste des flux chargés
		Feeds[feed.Id] = feed
		Feeds[feed.Id].Items = []*rss.Item{}

		c.Feeds[feed.Id] = feed

		if err != nil {
			return nil, err
		}

		return feed, nil
	}

	log.Println("Flux existant")
	oldFeed := Feeds[id]

	c.Feeds[id] = oldFeed

	return oldFeed, nil

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
		err := rows.Scan(&feed.Id, &feed.Nickname, &feed.Title, &feed.Description, &feed.Link, &feed.UpdateURL, &feed.Refresh, &feed.Unread, &feed.Insecure, &feed.Credentials.Username, &feed.Credentials.Password)
		if err != nil {
			return err
		}
		Feeds[feed.Id] = &feed
	}

	err = rows.Err()
	if err != nil {
		return err
	}

	return nil
}

func GetFeedArticles(feedId int64) (articles []rss.Item, err error) {

	var article rss.Item

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

func SaveArticle(id int64, c Context) (err error) {

	// On enregistre l'article dans la timeline des sauvegardes de l'utilsateur
	stmt2, err := db.Prepare("INSERT INTO article_timelines(article_id, timeline_id) VALUES(?, ?)")

	timelineId := c.Archive.Id
	_, err = stmt2.Exec(id, timelineId)
	if err != nil {
		return err
	}

	c.Archive.Size++

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

		// Insertion des articles dans les timelines
		for _, timelineId := range timelinesId {

			_, err = stmt2.Exec(lastId, timelineId)
			if err != nil {
				return err
			}

		}

	}

	// On met à jour les timelines en mémoire
	LoadTimelines()
	UpdateUsers()

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

	return nil

}

func SoftRemoveArticle(id int64, timelineName string, c Context) (err error) {

	sqlQ := "UPDATE article_timelines SET delete_date = NOW() WHERE article_id = ? AND ("
	var args []interface{}

	args = append(args, id)

	if timelineName == "global" {
		for _, timeline := range c.Timelines {
			sqlQ += "timeline_id = ? OR "
			args = append(args, timeline.Id)
		}
		sqlQ = sqlQ[:len(sqlQ)-3] + ")"
		UpdateUser(c.User.Id)

	} else if timelineName == "archive" {

		timeId := c.Archive.Id
		sqlQ += "timeline_id = ?) "
		args = append(args, timeId)
		c.Archive.Size--
		Archives[c.Archive.Id].Size--

	} else {
		tmp, _ := strconv.Atoi(timelineName)
		timeId := int64(tmp)
		sqlQ += "timeline_id = ?) "
		args = append(args, timeId)

		Timelines[timeId].Size--
		c.Timelines[timeId].Size--

	}

	stmt, err := db.Prepare(sqlQ)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(args...)
	if err != nil {
		return err
	}

	return nil

}

func RecoverArticle(id int64, timelineName string, c Context) (err error) {

	sqlQ := "UPDATE article_timelines SET delete_date = NULL WHERE article_id = ? AND ("
	var args []interface{}

	args = append(args, id)

	if timelineName == "global" {
		for _, timeline := range c.Timelines {
			sqlQ += "timeline_id = ? OR "
			args = append(args, timeline.Id)
		}
		sqlQ = sqlQ[:len(sqlQ)-3] + ")"
		UpdateUsers()

	} else if timelineName == "archive" {

		timeId := c.Archive.Id
		sqlQ += "timeline_id = ?) "
		args = append(args, timeId)
		c.Archive.Size++
		Archives[c.Archive.Id].Size++

	} else {
		tmp, _ := strconv.Atoi(timelineName)
		timeId := int64(tmp)
		sqlQ += "timeline_id = ?) "
		args = append(args, timeId)
		Timelines[timeId].Size++
		c.Timelines[timeId].Size++

	}

	stmt, err := db.Prepare(sqlQ)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(args...)
	if err != nil {
		return err
	}

	return nil

}

func RemoveArticle(id int64, timelineId int64) (err error) {

	sqlQ := "DELETE FROM article_timelines WHERE article_id = ? AND timeline_id = ?"
	var args []interface{}

	args = append(args, id)
	args = append(args, timelineId)

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

	return nil

}

func RemoveTimelineArticles(timelineName string, c Context) (err error) {

	sql := "DELETE FROM article_timelines WHERE ("
	var args []interface{}

	if timelineName == "global" {
		for _, timeline := range c.Timelines {
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
		c.Timelines[timeId].Size = 0

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
