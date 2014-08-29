package storage

type User struct {
	Id              int64
	Username        string
	Password        string
	SavedTimelineId int64
}

var Users []User
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
	CurrentUser = Users[0]

	err = rows.Err()
	if err != nil {
		return err
	}

	return nil

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
