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
var Archives map[int64]*Timeline

var UserTimelines map[int64]map[int64]*Timeline

func GetTimeline(name string, c Context) (t Timeline, err error) {

	if name == "archive" {
		err = db.QueryRow("select t.id, t.timeline, t.title, (SELECT COUNT(*) FROM article_timelines WHERE timeline_id = t.id AND delete_date IS NULL) size FROM timeline t  WHERE t.id = ?", c.Archive.Id).Scan(&t.Id, &t.Timeline, &t.Title, &t.Size)

		if err != nil && err != sql.ErrNoRows {
			return Timeline{}, err
		}

		if err == sql.ErrNoRows {
			return Timeline{}, nil
		}
		return
	}

	err = db.QueryRow("select t.id, t.timeline, t.title, (SELECT COUNT(*) FROM article_timelines WHERE timeline_id = t.id AND delete_date IS NULL) size, f.id, f.nickname, f.title, f.description, f.link, f.updateUrl, f.refresh, f.unread from timeline as t LEFT JOIN feed as f ON f.id = t.feed_id where t.id = ?", name).Scan(&t.Id, &t.Timeline, &t.Title, &t.Size, &t.Feed.Id, &t.Feed.Nickname, &t.Feed.Title, &t.Feed.Description, &t.Feed.Link, &t.Feed.UpdateURL, &t.Feed.Refresh, &t.Feed.Unread)

	if err != nil && err != sql.ErrNoRows {
		return Timeline{}, err
	}

	if err == sql.ErrNoRows {
		return Timeline{}, nil
	}

	return

}

func GetUserTimelines(userId int64) (ids []Timeline, err error) {

	rows, err := db.Query("SELECT id, feed_id FROM timeline WHERE user_id = ? ", userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var t Timeline
		err := rows.Scan(&t.Id, &t.Feed.Id)
		if err != nil {
			return nil, err
		}
		ids = append(ids, t)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return

}

func LoadTimelines() (err error) {

	Timelines = make(map[int64]*Timeline)
	Archives = make(map[int64]*Timeline)
	// Initialization de la map
	log.Println("Chargement des Timelines")

	rows, err := db.Query("select t.id, t.timeline, t.title, (SELECT COUNT(*) FROM article_timelines WHERE timeline_id = t.id AND delete_date IS NULL) size ,f.id, f.nickname, f.title, f.description, f.link, f.updateUrl, f.refresh, f.unread from timeline as t LEFT JOIN feed as f ON f.id = t.feed_id WHERE t.user_id IS NOT NULL")
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

	rows, err = db.Query("select t.id, t.timeline, t.title, (SELECT COUNT(*) FROM article_timelines WHERE timeline_id = t.id AND delete_date IS NULL) size FROM timeline t WHERE t.user_id IS NULL")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var t Timeline
		err := rows.Scan(&t.Id, &t.Timeline, &t.Title, &t.Size)
		if err != nil {
			return err
		}
		Archives[t.Id] = &t
	}
	err = rows.Err()
	if err != nil {
		return err
	}

	return nil

}

func CreateTimeline(title string, feed *rss.Feed, c Context) (err error) {

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

	res, err := stmt.Exec(title, title, 0, feed.Id, c.User.Id)
	if err != nil {
		return err
	}
	lastId, err := res.LastInsertId()

	// Assignation de l'id
	timeline.Id = lastId

	// Ajout à la liste des flux chargés
	Timelines[timeline.Id] = &timeline
	c.Timelines[timeline.Id] = &timeline

	if err != nil {
		return err
	}

	stmt.Close()
	return nil

}

func GetTimelineArticles(timelineId int64, nextId int64) (articles []rss.Item, err error) {

	var article rss.Item
	var articleFeedId int64

	var rows *sql.Rows

	if nextId == 0 {
		rows, err = db.Query("select a.id, a.date, a.description, a.link, a.pubdate, a.title, a.feed_id from article as a LEFT JOIN article_timelines as at ON a.id = at.article_id WHERE at.delete_date IS NULL AND at.timeline_id = ? ORDER BY a.pubdate ASC LIMIT ?", timelineId, MaxArticles)
	} else {
		rows, err = db.Query("select a.id, a.date, a.description, a.link, a.pubdate, a.title, a.feed_id from article as a LEFT JOIN article_timelines as at ON a.id = at.article_id WHERE at.delete_date IS NULL AND at.timeline_id = ? AND a.id != ? AND a.pubdate >= (SELECT pubdate FROM article WHERE id = ?) ORDER BY a.pubdate ASC LIMIT ?", timelineId, nextId, nextId, MaxArticles)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&article.Id, &article.Date, &article.Content, &article.Link, &article.PubDate, &article.Title, &articleFeedId)
		if err != nil {
			return nil, err
		}

		article.Feed = Feeds[articleFeedId].Title

		articles = append(articles, article)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return articles, nil

}

func GetGlobalArticles(nextId int64, c Context) (articles []rss.Item, err error) {

	var article rss.Item
	var args []interface{}
	var sql string

	// Création de la requête SQL
	if nextId == 0 {
		sql = "select a.id, a.date, a.description, a.link, a.pubdate, a.title, a.feed_id from article as a LEFT JOIN article_timelines as at ON a.id = at.article_id WHERE at.delete_date IS NULL AND ("
	} else {
		sql = "select a.id, a.date, a.description, a.link, a.pubdate, a.title, a.feed_id from article as a LEFT JOIN article_timelines as at ON a.id = at.article_id WHERE at.delete_date IS NULL AND a.id != ? AND a.pubdate >= (SELECT pubdate FROM article WHERE id = ?) AND ("
		args = append(args, nextId)
		args = append(args, nextId)
	}

	for _, timeline := range c.Timelines {
		sql += "at.timeline_id = ? OR "
		args = append(args, timeline.Id)
	}

	if len(c.Timelines) > 0 {
		sql = sql[:len(sql)-3]
	} else {
		return nil, nil
	}

	sql += ") ORDER BY a.pubdate ASC LIMIT ?"
	args = append(args, MaxArticles)

	rows, err := db.Query(sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articleFeedId int64

	for rows.Next() {
		err := rows.Scan(&article.Id, &article.Date, &article.Content, &article.Link, &article.PubDate, &article.Title, &articleFeedId)
		if err != nil {
			return nil, err
		}
		article.Feed = Feeds[articleFeedId].Title
		articles = append(articles, article)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return articles, nil

}

func GetGlobalArticlesSize(c Context) (size int, err error) {

	// Création de la requête SQL
	sql := "select COUNT(*) size from article as a LEFT JOIN article_timelines as at ON a.id = at.article_id WHERE at.delete_date IS NULL AND ("
	var args []interface{}

	for _, timeline := range c.Timelines {
		sql += "at.timeline_id = ? OR "
		args = append(args, timeline.Id)
	}

	if len(c.Timelines) > 0 {
		sql = sql[:len(sql)-3]
	} else {
		return 0, nil
	}
	sql += ")"
	err = db.QueryRow(sql, args...).Scan(&size)

	if err != nil {
		return 0, err
	}

	return

}

func RemoveTimeline(feedId int64, c Context) (err error) {

	var timelineId int64

	// Récupération de l'id de la timeleine
	err = db.QueryRow("select id FROM timeline WHERE feed_id = ? AND user_id = ?", feedId, c.User.Id).Scan(&timelineId)

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

	stmt.Close()

	// Suppression de la timeline
	stmt, err = db.Prepare("DELETE FROM timeline WHERE id = ?;")

	if err != nil {
		return err
	}

	_, err = stmt.Exec(timelineId)
	if err != nil {
		return err
	}

	stmt.Close()

	// On la supprime également de la mémoire
	delete(Timelines, timelineId)
	delete(c.Timelines, timelineId)
	delete(c.Feeds, feedId)

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
