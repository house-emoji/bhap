package main

import (
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

const minPasswordLength = 5

// serveLoginPage serves the page for logging in.
func serveLoginPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "views/login.html")
}

// handleLoginForm attempts to log the user in using credentials from a POST
// form.
func handleLoginForm(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	email := r.FormValue("email")
	password := r.FormValue("password")

	loggedIn, err := checkLogin(ctx, email, password)
	if err != nil {
		log.Errorf(ctx, "Error authenticating: %v", err)
		http.Error(w, "Error authenticating", http.StatusInternalServerError)
		return
	}
	if !loggedIn {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	loginSession, err := sessionStore.Get(r, "login")
	if err != nil {
		http.Error(w, "Could not decode session", http.StatusInternalServerError)
		log.Errorf(ctx, "could not decode session: %v", err)
		return
	}

	loginSession.Values["email"] = email

	if err := loginSession.Save(r, w); err != nil {
		http.Error(w, "Error logging in", http.StatusInternalServerError)
		log.Errorf(ctx, "could not save session: %v", err)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// logout logs the user out.
func logout(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	if err := deleteSession(w, r); err != nil {
		http.Error(w, "Could not log out", http.StatusInternalServerError)
		log.Errorf(ctx, "could not log out: %v", err)
		return
	}
}
