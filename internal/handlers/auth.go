package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"golang.org/x/crypto/bcrypt"
)

func (app *Application) Login(w http.ResponseWriter, r *http.Request) {
	files := []string{
		filepath.Join("ui", "html", "layouts", "auth.tmpl"),
		filepath.Join("ui", "html", "pages", "auth", "login.tmpl"),
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = ts.ExecuteTemplate(w, "auth", nil)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (app *Application) LoginPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	email := r.PostForm.Get("email")
	password := r.PostForm.Get("password")

	user, err := app.Models.GetUserByEmail(email)
	if err != nil {
		app.jsonResponse(w, http.StatusOK, map[string]string{"success": "false", "msg": "Invalid email or password"})
		return
	}

	// Check password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		app.jsonResponse(w, http.StatusOK, map[string]string{"success": "false", "msg": "Invalid email or password"})
		return
	}

	// Set session cookie (simple version)
	cookie := http.Cookie{
		Name:     "user_id",
		Value:    fmt.Sprintf("%d", user.ID),
		Path:     "/",
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)

	redirect := "/"
	if user.Role == "ShopKeeper" || user.Role == "StoreKeeper" {
		redirect = "/orders"
	}

	app.jsonResponse(w, http.StatusOK, map[string]string{
		"success":  "true",
		"msg":      "Login successful! Redirecting...",
		"redirect": redirect,
	})
}

func (app *Application) Logout(w http.ResponseWriter, r *http.Request) {
	// Clear the session cookie
	cookie := http.Cookie{
		Name:     "user_id",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	}
	http.SetCookie(w, &cookie)

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (app *Application) renderLoginWithError(w http.ResponseWriter, errMsg string) {
	files := []string{
		filepath.Join("ui", "html", "layouts", "auth.tmpl"),
		filepath.Join("ui", "html", "pages", "auth", "login.tmpl"),
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := struct {
		Error string
	}{
		Error: errMsg,
	}

	ts.ExecuteTemplate(w, "auth", data)
}
