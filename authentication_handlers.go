package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rs/xid"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
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

// serveInvitePage serves the page that is used to create new invitations to
// join the BHAP consortium.
func serveInvitePage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "views/invite.html")
}

// handleInvitationForm creates a new invitation based on form input from a
// POST request.
func handleInvitationForm(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	email := r.FormValue("email")

	newInvitation := invitation{
		Email:     email,
		UID:       xid.New().String(),
		EmailSent: false,
	}

	key := datastore.NewKey(ctx, InvitationEntityName, "", 0, nil)
	if _, err := datastore.Put(ctx, key, &newInvitation); err != nil {
		log.Errorf(ctx, "could not create invitation: %v", err)
		http.Error(w, "Could not create invitation", 500)
		return
	}

	log.Infof(ctx, "created a new invitation for %v", email)

	http.Redirect(w, r, "/invite", http.StatusSeeOther)
}

// serveNewUserPage serves the page that can be used to create a new user from
// an invitation.
func serveNewUserPage(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	// Get the invitation UID
	uid := mux.Vars(r)["uid"]

	invite, key, err := invitationByUID(ctx, uid)
	if err != nil {
		http.Error(w,
			"Error while getting invitation information",
			http.StatusInternalServerError)
		log.Errorf(ctx, "could not get invitation information: %v", err)
		return
	}

	if key == nil {
		http.Error(w,
			"Invalid invitation ID. Are you trying to pull a fast one?",
			http.StatusBadRequest)
		log.Warningf(ctx, "new user request made with invalid invitation ID")
		return
	}

	filler := newUserFiller{
		InvitationUID: uid,
		Email:         invite.Email,
	}
	showTemplate(ctx, w, newUserTemplate, filler)
}

// handleNewUserForm creates a new user based on form data from a POST request.
func handleNewUserForm(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	// Get the invitation UID
	uid := mux.Vars(r)["uid"]

	firstName := r.FormValue("firstName")
	lastName := r.FormValue("lastName")
	password := r.FormValue("password")

	if strings.TrimSpace(firstName) == "" {
		http.Error(w, "First name must not be empty", http.StatusBadRequest)
		log.Warningf(ctx, "Empty first name")
		return
	} else if strings.TrimSpace(lastName) == "" {
		http.Error(w, "Last name must not be empty", http.StatusBadRequest)
		log.Warningf(ctx, "Empty last name")
		return
	} else if len(password) < minPasswordLength {
		http.Error(w,
			fmt.Sprintf("Password must be longer than %v characters", minPasswordLength),
			http.StatusBadRequest)
		log.Warningf(ctx, "Empty last name")
		return
	}

	invite, inviteKey, err := invitationByUID(ctx, uid)
	if err != nil {
		http.Error(w,
			"Error while getting invitation information",
			http.StatusInternalServerError)
		log.Errorf(ctx, "could not get invitation information: %v", err)
		return
	}

	if inviteKey == nil {
		http.Error(w,
			"Invalid invitation ID. Are you trying to pull a fast one?",
			http.StatusBadRequest)
		log.Warningf(ctx, "new user request made with invalid invitation ID")
		return
	}

	// Hash the user's password
	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w,
			"Error hashing password",
			http.StatusInternalServerError)
		log.Errorf(ctx, "could not hash password: %v", err)
		return
	}

	newUser := user{
		FirstName:    firstName,
		LastName:     lastName,
		Email:        invite.Email,
		PasswordHash: passwordHash,
	}

	userKey := datastore.NewKey(ctx, userEntityName, "", 0, nil)
	if _, err := datastore.Put(ctx, userKey, &newUser); err != nil {
		http.Error(w,
			"Error saving new user",
			http.StatusInternalServerError)
		log.Errorf(ctx, "could not save new user: %v", err)
		return
	}

	// Delete the invitation so it can't be reused
	if err := datastore.Delete(ctx, inviteKey); err != nil {
		log.Errorf(ctx, "could not delete used invitation: %v", err)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// requireLogin is middleware that requires the user be logged in.
func requireLogin(next func(w http.ResponseWriter, r *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := appengine.NewContext(r)

		loginSession, err := sessionStore.Get(r, "login")
		if err != nil {
			http.Error(w, "Could not decode session", http.StatusInternalServerError)
			log.Errorf(ctx, "could not decode session: %v", err)
			return
		}

		if loginSession.IsNew {
			log.Infof(ctx, "No session exists, redirecting to login")
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		next(w, r)
	})
}
