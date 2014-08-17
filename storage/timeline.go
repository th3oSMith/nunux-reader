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

var Timelines map[int64]*Timeline
var Archive *Timeline

func GetTimeline(name string) (t Timeline, err error) {

	if name == "archive" {
		err = db.QueryRow("select t.id, t.timeline, t.title, (SELECT COUNT(*) FROM article_timelines WHERE timeline_id = t.id) size FROM timeline t  WHERE t.id = ?", Archive.Id).Scan(&t.Id, &t.Timeline, &t.Title, &t.Size)

		if err != nil && err != sql.ErrNoRows {
			return Timeline{}, err
		}

		if err == sql.ErrNoRows {
			return Timeline{}, nil
		}
		log.Println("Récupération du statut des archives")
		return
	}

	log.Println("Récupération de la timeline")

	err = db.QueryRow("select t.id, t.timeline, t.title, (SELECT COUNT(*) FROM article_timelines WHERE timeline_id = t.id) size, f.id, f.nickname, f.title, f.description, f.link, f.updateUrl, f.refresh, f.unread from timeline as t LEFT JOIN feed as f ON f.id = t.feed_id where t.id = ?", name).Scan(&t.Id, &t.Timeline, &t.Title, &t.Size, &t.Feed.Id, &t.Feed.Nickname, &t.Feed.Title, &t.Feed.Description, &t.Feed.Link, &t.Feed.UpdateURL, &t.Feed.Refresh, &t.Feed.Unread)

	if err != nil && err != sql.ErrNoRows {
		return Timeline{}, err
	}

	if err == sql.ErrNoRows {
		return Timeline{}, nil
	}

	return

}

func LoadTimelines() (err error) {

	Timelines = make(map[int64]*Timeline)
	// Initialization de la map
	log.Println("Chargement des Timelines")

	rows, err := db.Query("select t.id, t.timeline, t.title, (SELECT COUNT(*) FROM article_timelines WHERE timeline_id = t.id) size ,f.id, f.nickname, f.title, f.description, f.link, f.updateUrl, f.refresh, f.unread from timeline as t LEFT JOIN feed as f ON f.id = t.feed_id WHERE t.user_id IS NOT NULL")
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
		Timelines[t.Id] = &t
	}
	err = rows.Err()
	if err != nil {
		return err
	}

	var t Timeline
	err = db.QueryRow("select t.id, t.timeline, t.title, (SELECT COUNT(*) FROM article_timelines WHERE timeline_id = t.id) size FROM timeline t WHERE t.id = ?", CurrentUser.SavedTimelineId).Scan(&t.Id, &t.Timeline, &t.Title, &t.Size)

	if err != nil {
		return err
	}

	Archive = &t

	log.Println("Timelines chargées")
	log.Println(Timelines)

	return nil

}

func CreateTimeline(title string, feed *rss.Feed) (err error) {

	var timeline Timeline

	timeline.Timeline = title
	timeline.Title = title
	timeline.Size = 0
	timeline.Feed = *feed

	// Enregistrement dans la base SQL
	stmt, err := db.Prepare("INSERT INTO timeline(timeline, title, size, feed_id, user_id) VALUES(?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}

	log.Println("userId", CurrentUser.Id)

	res, err := stmt.Exec(title, title, 0, feed.Id, CurrentUser.Id)
	if err != nil {
		return err
	}
	lastId, err := res.LastInsertId()

	// Assignation de l'id
	timeline.Id = lastId

	// Ajout à la liste des flux chargés
	Timelines[timeline.Id] = &timeline

	if err != nil {
		return err
	}
	rowCnt, err := res.RowsAffected()

	if err != nil {
		return err
	}
	log.Printf("Insertion d'une Timeline ID = %d, affected = %d\n", lastId, rowCnt)

	return nil

}

func GetTimelineArticles(timelineId int64) (articles []rss.Item, err error) {

	var article rss.Item

	log.Println("Chargement des articles de la Timeline", timelineId)

	rows, err := db.Query("select a.id, a.date, a.description, a.link, a.pubdate, a.title from article as a LEFT JOIN article_timelines as at ON a.id = at.article_id WHERE at.timeline_id = ? ORDER BY a.pubdate ASC", timelineId)
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

func GetGlobalArticles() (articles []rss.Item, err error) {

	var article rss.Item

	log.Println("Chargement des articles de la Timeline Globale")

	// Création de la requête SQL
	sql := "select a.id, a.date, a.description, a.link, a.pubdate, a.title from article as a LEFT JOIN article_timelines as at ON a.id = at.article_id WHERE "
	var args []interface{}

	for _, timeline := range Timelines {
		sql += "at.timeline_id = ? OR "
		args = append(args, timeline.Id)
	}

	if len(Timelines) > 0 {
		sql = sql[:len(sql)-3]
	} else {
		sql = sql[:len(sql)-6]
	}

	sql += " ORDER BY a.pubdate ASC"

	log.Println(sql)

	rows, err := db.Query(sql, args...)
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

func GetGlobalArticlesSize() (size int, err error) {

	log.Println("Chargement des articles de la Timeline Globable")

	// Création de la requête SQL
	sql := "select COUNT(*) size from article as a LEFT JOIN article_timelines as at ON a.id = at.article_id WHERE "
	var args []interface{}

	for _, timeline := range Timelines {
		sql += "at.timeline_id = ? OR "
		args = append(args, timeline.Id)
	}

	if len(Timelines) > 0 {
		sql = sql[:len(sql)-3]
	} else {
		sql = sql[:len(sql)-6]
	}

	err = db.QueryRow(sql, args...).Scan(&size)

	if err != nil {
		return 0, err
	}

	return

}

func RemoveTimeline(feedId int64) (err error) {

	var timelineId int64

	// Récupération de l'id de la timeleine
	err = db.QueryRow("select id FROM timeline WHERE feed_id = ? AND user_id = ?", feedId, CurrentUser.Id).Scan(&timelineId)

	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if err == sql.ErrNoRows {
		return nil
	}

	// Suppression des articles de la timeline
	stmt, err := db.Prepare("DELETE FROM article_timelines WHERE timeline_id = ?;")

	if err != nil {
		return err
	}

	_, err = stmt.Exec(timelineId)
	if err != nil {
		return err
	}

	// Suppression de la timeline
	stmt, err = db.Prepare("DELETE FROM timeline WHERE id = ?;")

	if err != nil {
		return err
	}

	_, err = stmt.Exec(timelineId)
	if err != nil {
		return err
	}

	log.Printf("Suppression de la timeline ID = %d", feedId)

	// On la supprime également de la mémoire
	delete(Timelines, timelineId)

	// On supprime le flux si jamais il n'est plus utilisé

	var number int64

	// Récupération du nombre de timelines l'utilisant
	err = db.QueryRow("select COUNT(*) number FROM timeline WHERE feed_id = ?", feedId).Scan(&number)

	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if err == sql.ErrNoRows {
		return nil
	}

	if number == 0 {
		RemoveFeed(feedId)
	}

	return nil

}
