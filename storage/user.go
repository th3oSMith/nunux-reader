package storage

type User struct {
	Id       int64
	Username string
	Password string
}

var Users []User

func LoadUsers() (err error) {

	var id int64
	var username, password string

	rows, err := db.Query("select * FROM user;")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&id, &username, &password)
		if err != nil {
			return err
		}
		Users = append(Users, User{id, username, password})
	}
	err = rows.Err()
	if err != nil {
		return err
	}

	return nil

}
