package storage

import (
	"github.com/th3osmith/rss"
	"log"
)

type User struct {
	Id              int64  `json:"id"`
	Username        string `json:"username"`
	Password        string `json:"password"`
	SavedTimelineId int64  `json:"-"`
}

var Users map[int64]User
var CurrentUsers map[string]User
var CurrentUser User

func LoadUsers() (err error) {

	var id, savedTimelineId int64
	var username, password string
	Users = make(map[int64]User)

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
		Users[id] = User{id, username, password, savedTimelineId}
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

			if len(UserFeeds[user.Id][t.Feed.Id].Credentials.Password) > 0 {
				plainPwd, _ := Decrypt(UserFeeds[user.Id][t.Feed.Id].Credentials.Password, user.Password)
				UserFeeds[user.Id][t.Feed.Id].Credentials.Password = plainPwd
				Feeds[t.Feed.Id].Credentials.Password = plainPwd

			}
			log.Println("Pwd", UserFeeds[user.Id][t.Feed.Id].Credentials.Password)
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

func CreateUser(user User) (createdUser User, err error) {

	// Création de la timeline d'articles sauvegardés de l'utilisateur
	stmt, err := db.Prepare("INSERT INTO timeline(timeline, title) VALUES(?, ?)")
	if err != nil {
		return User{}, err
	}

	res, err := stmt.Exec("archive", "Saved Articles")
	if err != nil {
		return User{}, err
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		return User{}, err
	}

	// Insertion de l'utilisateur
	stmt, err = db.Prepare("INSERT INTO user(username, password, saved_timeline_id) VALUES(?, ?, ?)")

	if err != nil {
		return User{}, err
	}

	res, err = stmt.Exec(user.Username, user.Password, lastId)
	if err != nil {
		return User{}, err
	}

	lastId, err = res.LastInsertId()
	if err != nil {
		return User{}, err
	}

	user.Id = lastId
	return user, nil
}

func DeleteUser(userId int64) (err error) {

	// On supprime les timelines
	timelines := UserTimelines[userId]

	c := Context{}
	c.User = Users[userId]
	c.Feeds = UserFeeds[userId]
	c.Timelines = UserTimelines[userId]
	c.Archive = Archives[c.User.SavedTimelineId]

	for _, timeline := range timelines {
		RemoveTimeline(timeline.Feed.Id, c)
	}

	// On supprime l'utilisateur et sa timeline archive
	stmt, err := db.Prepare("DELETE FROM user WHERE id = ?;")

	if err != nil {
		return err
	}

	_, err = stmt.Exec(c.User.Id)
	if err != nil {
		return err
	}
	delete(Users, c.User.Id)

	stmt, err = db.Prepare("DELETE FROM timeline WHERE id = ?;")

	if err != nil {
		return err
	}

	_, err = stmt.Exec(c.Archive.Id)
	if err != nil {
		return err
	}
	delete(Archives, c.Archive.Id)
	delete(UserTimelines, c.User.Id)
	delete(UserFeeds, c.User.Id)

	return nil

}

func UpdateUserInformations(user User) (err error) {

	stmt, err := db.Prepare("UPDATE user SET username = ?, password = ? WHERE id = ?;")

	if err != nil {
		return err
	}

	_, err = stmt.Exec(user.Username, Sha256Sum(user.Password), user.Id)
	if err != nil {
		return err
	}

	Users[user.Id] = user
	return nil

}
