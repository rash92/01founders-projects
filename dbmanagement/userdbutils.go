package dbmanagement

import (
	"database/sql"
	"errors"
	"fmt"
	"forum/utils"
)

/*
Generates a new user in the database.  The UUID is generated internally here and stored to the database (this can also be referred to as the userID).

The inserted User is also returned in case it is needed to be used straight away but it is not necessary.
*/
func InsertUser(name string, email string, password string, permission string) User {
	db, _ := sql.Open("sqlite3", "./forum.db")
	defer db.Close()
	utils.WriteMessageToLogFile("Inserting user record...")

	UUID := GenerateUUIDString()
	tokens := 3
	insertUserData := "INSERT INTO Users(UUID, name, email, password, permission, limitTokens) VALUES (?, ?, ?, ?, ?, ?)"
	statement, err := db.Prepare(insertUserData)
	utils.HandleError("User Prepare failed in InserUser function", err)

	_, err = statement.Exec(UUID, name, email, password, permission, tokens)
	utils.HandleError("Statement Exec failed in InserUser function", err)

	return User{UUID, name, email, password, permission, []Notification{}, tokens}
}

func UpdateUserPermissionFromUUID(UUID string, newpermission string) {
	db, _ := sql.Open("sqlite3", "./forum.db")
	defer db.Close()
	message := fmt.Sprintf("Updating user permission to: %v", newpermission)
	utils.WriteMessageToLogFile(message)

	updateUserData := "UPDATE Users SET permission = ? WHERE uuid = ?"
	statement, err := db.Prepare(updateUserData)
	utils.HandleError("User Update Prepare failed in UpdateUserPermissionFromUUID function", err)

	_, err = statement.Exec(newpermission, UUID)
	utils.HandleError("Statement Exec failed in UpdateUserPermissionFromUUID function", err)
}

func UpdateUserPermissionFromName(Name string, newpermission string) {
	db, _ := sql.Open("sqlite3", "./forum.db")
	defer db.Close()
	message := fmt.Sprintf("Updating user permission to: %v", newpermission)
	utils.WriteMessageToLogFile(message)

	updateUserData := "UPDATE Users SET permission = ? WHERE name = ?"
	statement, err := db.Prepare(updateUserData)
	utils.HandleError("User Update Prepare failed", err)

	_, err = statement.Exec(newpermission, Name)
	utils.HandleError("Statement Exec failed in", err)
}

/*
Generates a new user in the database.  The UUID is generated internally here and stored to the database (this can also be referred to as the userID).

The inserted User is also returned in case it is needed to be used straight away but it is not necessary.
*/
func DeleteUser(name string) error {
	db, _ := sql.Open("sqlite3", "./forum.db")
	defer db.Close()

	statement := "DELETE FROM Users WHERE name = ?"
	stm, err := db.Prepare(statement)
	utils.HandleError("Failed to delete user statement in", err)

	defer stm.Close()

	res, err := stm.Exec(name)
	utils.HandleError("Failed to delete user in", err)

	n, err := res.RowsAffected()
	utils.HandleError("Rows affected error in", err)

	message := fmt.Sprintf("The statement has affected %d rows\n", n)
	utils.WriteMessageToLogFile(message)
	return err
}

/*
Used to display all currently registered users.  Should only be used internally as information is not relevant for the website.
*/
func DisplayAllUsers() {
	db, _ := sql.Open("sqlite3", "./forum.db")
	defer db.Close()

	row, err := db.Query("SELECT * FROM Users ORDER BY name")
	utils.HandleError("User query failed in", err)
	defer row.Close()

	for row.Next() {

		var UUID string
		var name string
		var email string
		var password string
		var permission string
		var tokens int
		row.Scan(&UUID, &name, &email, &password, &permission, &tokens)
		message := fmt.Sprint("User: ", UUID, " ", name, " ", email, " ", password, " ", permission)
		utils.WriteMessageToLogFile(message)
	}
}

func SelectAllUsers() []User {
	db, _ := sql.Open("sqlite3", "./forum.db")
	defer db.Close()

	row, err := db.Query("SELECT * FROM Users")
	utils.HandleError("User query failed", err)
	defer row.Close()

	var allUsers []User
	for row.Next() {
		var currentUser User
		row.Scan(&currentUser.UUID, &currentUser.Name, &currentUser.Email, &currentUser.Password, &currentUser.Permission, &currentUser.LimitTokens)
		allUsers = append(allUsers, currentUser)
	}
	return allUsers
}

/*
Initially used for when a user is trying to log in.  Returns a User's information when searched for by name.
*/
func SelectUserFromName(Name string) (User, error) {
	var user User
	db, _ := sql.Open("sqlite3", "./forum.db")
	defer db.Close()

	stm, err := db.Prepare("SELECT * FROM Users WHERE name = ?")
	utils.HandleError("Statement failed in", err)

	err = stm.QueryRow(Name).Scan(&user.UUID, &user.Name, &user.Email, &user.Password, &user.Permission, &user.LimitTokens)
	utils.HandleError("Query Row failed in", err)

	return user, err
}

/*
Could be used for if a user wanted to log in using their email address.  Returns a User's information when searched for by email.
*/
func SelectUserFromEmail(Email string) (User, error) {
	var user User
	db, _ := sql.Open("sqlite3", "./forum.db")
	defer db.Close()

	stm, err := db.Prepare("SELECT * FROM Users WHERE email = ?")
	utils.HandleError("Statement failed", err)

	err = stm.QueryRow(Email).Scan(&user.UUID, &user.Name, &user.Email, &user.Password, &user.Permission, &user.LimitTokens)
	utils.HandleError("Query Row failed", err)

	return user, err
}

/*
Used when you have the users UUID (userID).  For example, within a session (displaying user information such as username), or when displaying post and comment details.
*/
func SelectUserFromUUID(UUID string) (User, error) {
	var user User
	db, _ := sql.Open("sqlite3", "./forum.db")
	defer db.Close()

	stm, err := db.Prepare("SELECT * FROM Users WHERE uuid = ?")
	utils.HandleError("Statement failed", err)

	err = stm.QueryRow(UUID).Scan(&user.UUID, &user.Name, &user.Email, &user.Password, &user.Permission, &user.LimitTokens)
	utils.HandleError("Query Row failed", err)

	return user, err
}

/*
Gets the user using the current session.  Used to assign the correct userID if a user posts, likes, dislikes, or comments.
*/
func SelectUserFromSession(UUID string) (User, error) {
	db, _ := sql.Open("sqlite3", "./forum.db")
	defer db.Close()

	var userID string
	err := db.QueryRow("SELECT userId FROM Sessions WHERE uuid = ?", UUID).Scan(&userID)
	utils.HandleError("User from session query failed", err)

	var user User
	err = db.QueryRow("SELECT * FROM Users WHERE uuid = ?", userID).Scan(&user.UUID, &user.Name, &user.Email, &user.Password, &user.Permission, &user.LimitTokens)
	utils.HandleError("User query failed", err)

	return user, err
}

func UpdateUserToken(UUID string, n int) error {
	usertoken := GetUserToken(UUID)
	if usertoken == 0 && n != 3 {
		utils.WriteMessageToLogFile("Token limit reached for user")
		return errors.New("limit reached")
	} else {

		var tokenStatement string

		if n == 3 {
			tokenStatement = `
	UPDATE Users 
	SET limitTokens = ?
	WHERE uuid = ?
`
		} else {
			tokenStatement = `
	UPDATE Users 
	SET limitTokens = limitTokens - ?
	WHERE uuid = ?
`
		}

		db, _ := sql.Open("sqlite3", "./forum.db")
		defer db.Close()

		statement, err := db.Prepare(tokenStatement)
		utils.HandleError("token statement failed", err)

		_, err = statement.Exec(n, UUID)
		utils.HandleError("token statement Exec failed", err)

		return err
	}
}

func GetUserToken(UUID string) int {
	var user User
	db, _ := sql.Open("sqlite3", "./forum.db")
	defer db.Close()

	stm, err := db.Prepare("SELECT * FROM Users WHERE uuid = ?")
	utils.HandleError("Statement failed", err)

	err = stm.QueryRow(UUID).Scan(&user.UUID, &user.Name, &user.Email, &user.Password, &user.Permission, &user.LimitTokens)
	utils.HandleError("Query Row failed", err)

	return user.LimitTokens
}
