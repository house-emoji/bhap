package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

const minPasswordLength = 5

// newUserFiller fills the new user sign-up page template.
type newUserFiller struct {
	InvitationUID string
	Email         string
	BackgroundURL string
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

	backgroundURL, err := randomBackgroundURL()
	if err != nil {
		http.Error(w, "Error while looking for backgrounds",
			http.StatusInternalServerError)
		log.Errorf(ctx, "looking for backgrounds: %v", err)
		return
	}

	filler := newUserFiller{
		InvitationUID: uid,
		Email:         invite.Email,
		BackgroundURL: backgroundURL,
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
