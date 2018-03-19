package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

const minPasswordLength = 5

// createBHAP creates a new BHAP based on information passed from a POST form.
func createBHAP(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	title := r.FormValue("title")
	content := r.FormValue("content")

	// Find what the ID of the new BHAP should be
	// TODO(velovix): Nasty race condition here. Some kind of database lock
	// should fix this
	newID, err := nextID(ctx)
	if err != nil {
		log.Errorf(ctx, "could not query BHAPs: %v", err)
		http.Error(w, "Error while finding BHAP", http.StatusInternalServerError)
		return
	}

	_, userKey, err := userFromSession(ctx, r)
	if err != nil {
		http.Error(w, "Could not read session", http.StatusInternalServerError)
		log.Errorf(ctx, "could not get session email: %v", err)
		return
	}

	newBHAP := bhap{
		ID:           newID,
		Title:        title,
		LastModified: time.Now(),
		Author:       userKey,
		Status:       draftStatus,
		CreatedDate:  time.Now(),
		Content:      content,
	}

	// Save the new BHAP
	key := datastore.NewKey(ctx, "BHAP", "", 0, nil)
	if _, err := datastore.Put(ctx, key, &newBHAP); err != nil {
		log.Errorf(ctx, "failed to save BHAP: %v", err)
		http.Error(w, "Could not save BHAP", http.StatusInternalServerError)
		return
	}

	log.Infof(ctx, "Saved BHAP %v: %v", newID, title)

	http.Redirect(w, r, "/bhap/"+string(newID), http.StatusFound)
}

func newUser(w http.ResponseWriter, r *http.Request) {
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

	userKey := datastore.NewKey(ctx, UserEntityName, "", 0, nil)
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
}
