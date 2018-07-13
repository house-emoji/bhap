package main

import (
	"io/ioutil"
	"math/rand"
	"net/http"
	"path"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

var loginTemplate = compileTempl("views/login.html")

type loginPageFiller struct {
	BackgroundURL string
}

// serveLoginPage serves the page for logging in.
func serveLoginPage(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	backgroundURL, err := randomBackgroundURL()
	if err != nil {
		http.Error(w, "Error while looking for backgrounds",
			http.StatusInternalServerError)
		log.Errorf(ctx, "looking for backgrounds: %v", err)
		return
	}

	filler := loginPageFiller{
		BackgroundURL: backgroundURL,
	}

	showTemplate(ctx, w, loginTemplate, filler)
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

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func randomBackgroundURL() (string, error) {
	const backgroundPath = "static/backgrounds"

	backgrounds, err := ioutil.ReadDir(backgroundPath)
	if err != nil {
		return "", err
	}

	backgroundName := backgrounds[rand.Intn(len(backgrounds))]

	return path.Join("/", backgroundPath, backgroundName.Name()), nil
}
