package auth

import (
	"forum/dbmanagement"
	"forum/utils"
	"html/template"
	"log"
	"net/http"
)

type OauthAccount struct {
	Name, Email string
}

// Displays the log in page.
func Login(w http.ResponseWriter, r *http.Request, tmpl *template.Template) {
	sessionId, err := GetSessionFromBrowser(w, r)
	utils.HandleError("Unable to get session id from browser in login:", err)
	_, err = dbmanagement.SelectUserFromSession(sessionId)
	if err == nil {
		http.Redirect(w, r, "/", http.StatusFound)
	} else {
		tmpl.ExecuteTemplate(w, "login.html", nil)
	}
}

/*
Authenticate user with credentials - If the username and password match an entry in the database then the user is redirected to the forum page,
otherwise the user stays on the log in page. Session Cookie is also set here.
*/
func Authenticate(w http.ResponseWriter, r *http.Request, tmpl *template.Template) {
	if r.Method == "POST" {
		userName := r.FormValue("user_name")
		password := r.FormValue("password")

		user, err := dbmanagement.SelectUserFromName(userName)
		utils.HandleError("unable to get user error:", err)

		if CompareHash(user.Password, password) {
			err := CreateUserSession(w, r, user)
			utils.HandleError("Failed to create session in authenticate", err)
			http.Redirect(w, r, "/forum", http.StatusSeeOther)
		} else {
			log.Println("Incorrect Password!")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
		}
	}
}

// Logs user out
func Logout(w http.ResponseWriter, r *http.Request, tmpl *template.Template) {
	log.Println("logging out...")
	cookie, err := r.Cookie("session")
	utils.HandleError("Failed to get cookie", err)
	if err != http.ErrNoCookie {
		session := cookie.Value
		err := dbmanagement.DeleteSessionByUUID(session)
		utils.HandleError("Failed to get cookie", err)
	}
	clearcookie := http.Cookie{
		Name:     "session",
		Value:    "",
		HttpOnly: true,
		Path:     "/",
	}
	http.SetCookie(w, &clearcookie)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Displays the register page
func Register(w http.ResponseWriter, r *http.Request, tmpl *template.Template) {
	tmpl.ExecuteTemplate(w, "register.html", nil)
}

// Registers a user with given details then redirects to log in page.  Password is hashed here.
func RegisterAcount(w http.ResponseWriter, r *http.Request, tmpl *template.Template) {
	if r.Method == "POST" {
		userName := r.FormValue("user_name")
		email := r.FormValue("email")
		password := HashPassword(r.FormValue("password"))
		dbmanagement.InsertUser(userName, email, password, "user")
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
