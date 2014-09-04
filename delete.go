package main

import (
	"github.com/th3osmith/nunux-reader/storage"
	"log"
	"time"
)

var del *deleter

type deleter struct {
	quit chan struct{}
}

func InitDeleter() {
	del = NewDeleter()
	go del.Run()
}

func NewDeleter() *deleter {
	out := new(deleter)
	out.quit = make(chan struct{})
	return out
}

func (u *deleter) Run() {
	ticker := time.NewTicker(time.Duration(storage.DeleteTime) * time.Minute)
	Delete()

	for {
		select {
		case <-ticker.C:
			Delete()
		case <-u.quit:
			ticker.Stop()
			return
		}
	}
}

func Delete() (err error) {

	log.Println("Suppression des articles périmés")

	var articleId, timelineId int64

	rows, err := db.Query("SELECT article_id, timeline_id FROM article_timelines WHERE delete_date < NOW() - INTERVAL 1 WEEK")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&articleId, &timelineId)
		if err != nil {
			return err
		}
		storage.RemoveArticle(articleId, timelineId)
	}
	err = rows.Err()
	if err != nil {
		return err
	}

	return err

}

func (u *deleter) Quit() {
	close(u.quit)
}
