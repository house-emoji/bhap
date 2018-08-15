package pages

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/house-emoji/bhap"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

var bhapEditTemplate = compileTempl("views/edit.html")

type editPageFiller struct {
	LoggedIn bool
	FullName string
	BHAP     bhap.BHAP
}

// ServeEditPage serves up a page that allows the user to edit a proposal.
func ServeEditPage(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	// Load the BHAP
	loadedBHAP, bhapKey, err := bhapFromURLVars(ctx, mux.Vars(r))
	if err != nil {
		log.Errorf(ctx, "could not load BHAP: %v", err)
		http.Error(w, "Failed to load BHAP", http.StatusInternalServerError)
		return
	}
	if bhapKey == nil {
		http.Error(w, "No BHAP with that identifier", 404)
		log.Warningf(ctx, "unknown BHAP requested")
		return
	}

	if !isEditableStatus(loadedBHAP.Status) {
		http.Error(w, "Only draft or discussion BHAPs may be edited",
			http.StatusBadRequest)
		log.Warningf(ctx, "request to edit a non-draft or non-discussion proposal")
		return
	}

	var author bhap.User
	if err := datastore.Get(ctx, loadedBHAP.Author, &author); err != nil {
		log.Errorf(ctx, "Error loading user: %v", err)
		http.Error(w, "Failed to load user", http.StatusInternalServerError)
		return
	}

	// Get the current logged in user
	currUser, userKey, err := bhap.UserFromSession(ctx, r)
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

// HandleEdit handles a request to edit a BHAP.
func HandleEdit(op bhapOperator, w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	title := r.FormValue("title")
	shortDescription := r.FormValue("shortDescription")
	content := r.FormValue("content")

	if !op.bhap.Author.Equal(op.userKey) {
		http.Error(w, "Only authors may edit a BHAP",
			http.StatusForbidden)
		log.Warningf(ctx, "request from non-author denied")
		return
	}

	if !isEditableStatus(op.bhap.Status) {
		http.Error(w, "Only draft or discussion BHAPs may be edited",
			http.StatusBadRequest)
		log.Warningf(ctx, "request to edit a non-draft or non-discussion proposal")
		return
	}

	op.bhap.Title = title
	op.bhap.ShortDescription = shortDescription
	op.bhap.Content = content

	if _, err := datastore.Put(ctx, op.bhapKey, &op.bhap); err != nil {
		log.Errorf(ctx, "failed to update BHAP: %v", err)
		http.Error(w, "Could not update BHAP", http.StatusInternalServerError)
		return
	}

	if op.bhap.Status == bhap.DraftStatus {
		http.Redirect(w, r, fmt.Sprintf("/draft/%v", op.bhap.DraftID), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/bhap/%v", op.bhap.ID), http.StatusSeeOther)
	}
}

func isEditableStatus(status bhap.Status) bool {
	return status == bhap.DraftStatus || status == bhap.DiscussionStatus
}
