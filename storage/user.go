package storage

import (
	"github.com/th3osmith/rss"
	"log"
)

type User struct {
	Id              int64
	Username        string
	Password        string
	SavedTimelineId int64
}

var Users []User
var CurrentUsers map[string]User
var CurrentUser User

func LoadUsers() (err error) {

	var id, savedTimelineId int64
	var username, password string

	rows, err := db.Query("select * FROM user;")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&id, &username, &password, &savedTimelineId)
		if err != nil {
			return err
		}
		Users = append(Users, User{id, username, password, savedTimelineId})
	}

	// DEBUG --> User fixé
	//CurrentUser = Users[0]

	err = rows.Err()
	if err != nil {
		return err
	}

	return nil

}

func InitUser(user User) {

	// Si l'utilisateur n'est pas en mémoire on le récupère
	if len(UserFeeds[user.Id]) == 0 {

		log.Println("Initialisation du contexte de l'utilisateur")

		timelinesIds, err := GetUserTimelines(user.Id)
		if err != nil {
			log.Println("Impossible de créer le contexte de l'utilisateur")
			log.Println(err)
		}

		UserTimelines[user.Id] = make(map[int64]*Timeline)
		UserFeeds[user.Id] = make(map[int64]*rss.Feed)

		for _, t := range timelinesIds {
			UserTimelines[user.Id][t.Id] = Timelines[t.Id]
			UserFeeds[user.Id][t.Feed.Id] = Feeds[t.Feed.Id]
		}
	}

}

func UpdateUsers() {

	for userId, userTimelines := range UserTimelines {
		for id, _ := range userTimelines {
			UserTimelines[userId][id] = Timelines[id]
			feedId := UserTimelines[userId][id].Feed.Id
			UserFeeds[userId][feedId] = Feeds[feedId]
		}
	}
}

func UpdateUser(userId int64) {

	for id, _ := range UserTimelines[userId] {
		UserTimelines[userId][id] = Timelines[id]
		feedId := UserTimelines[userId][id].Feed.Id
		UserFeeds[userId][feedId] = Feeds[feedId]
	}

}

func CreateUser(username string, password string) (err error) {

	// Insertion de l'utilisateur
	stmt, err := db.Prepare("INSERT INTO user(username, password) VALUES(?, ?)")

	if err != nil {
		return err
	}

	_, err = stmt.Exec(username, password)
	if err != nil {
		return err
	}

	return nil
}
