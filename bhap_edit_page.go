package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

var bhapEditTemplate = compileTempl("views/edit.html")

type editPageFiller struct {
	LoggedIn bool
	FullName string
	BHAP     bhap
}

// serveBHAPEditPage serves up a page that allows the user to edit a proposal.
func serveBHAPEditPage(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	// Get the requested ID
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Warningf(ctx, "invalid ID %v: %v", idStr, err)
		http.Error(w, "ID must be an integer", 400)
		return
	}

	// Load the requested BHAP
	loadedBHAP, bhapKey, err := bhapByID(ctx, id)
	if err != nil {
		log.Errorf(ctx, "could not load BHAP: %v", err)
		http.Error(w, "Failed to load BHAP", http.StatusInternalServerError)
		return
	}
	if bhapKey == nil {
		http.Error(w, fmt.Sprintf("No BHAP with ID %v", id), 404)
		log.Warningf(ctx, "unknown BHAP requested: %v", id)
		return
	}

	if !isEditable(loadedBHAP.Status) {
		http.Error(w, "Only draft or discussion BHAPs may be edited",
			http.StatusBadRequest)
		log.Warningf(ctx, "request to edit a non-draft or non-discussion proposal")
		return
	}

	var author user
	if err := datastore.Get(ctx, loadedBHAP.Author, &author); err != nil {
		log.Errorf(ctx, "Error loading user: %v", err)
		http.Error(w, "Failed to load user", http.StatusInternalServerError)
		return
	}

	// Get the current logged in user
	currUser, userKey, err := userFromSession(ctx, r)
	if err != nil {
		http.Error(w, "Could not read session", http.StatusInternalServerError)
		log.Errorf(ctx, "could not get session email: %v", err)
		return
	}

	if !loadedBHAP.Author.Equal(userKey) {
		http.Error(w, "Only authors may edit a BHAP", http.StatusForbidden)
		log.Warningf(ctx, "request to edit by a non-author")
		return
	}

	filler := editPageFiller{
		LoggedIn: userKey != nil,
		FullName: currUser.FirstName + " " + currUser.LastName,
		BHAP:     loadedBHAP,
	}
	showTemplate(ctx, w, bhapEditTemplate, filler)
}

func handleEdit(op bhapOperator, w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	title := r.FormValue("title")
	content := r.FormValue("content")

	if !op.bhap.Author.Equal(op.userKey) {
		http.Error(w, "Only authors may edit a BHAP",
			http.StatusForbidden)
		log.Warningf(ctx, "request from non-author denied")
		return
	}

	if !isEditable(op.bhap.Status) {
		http.Error(w, "Only draft or discussion BHAPs may be edited",
			http.StatusBadRequest)
		log.Warningf(ctx, "request to edit a non-draft or non-discussion proposal")
		return
	}

	op.bhap.Title = title
	op.bhap.Content = content

	if _, err := datastore.Put(ctx, op.bhapKey, &op.bhap); err != nil {
		log.Errorf(ctx, "failed to update BHAP: %v", err)
		http.Error(w, "Could not update BHAP", http.StatusInternalServerError)
		return
	}

	log.Infof(ctx, "Updated BHAP %v:, %v", op.bhap.ID, op.bhap.Title)

	http.Redirect(w, r, fmt.Sprintf("/bhap/%v", op.bhap.ID), http.StatusSeeOther)
}

func isEditable(bhapStatus status) bool {
	return bhapStatus == draftStatus || bhapStatus == discussionStatus
}
