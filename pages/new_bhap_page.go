package pages

import (
	"fmt"
	"net/http"
	"time"

	"github.com/house-emoji/bhap"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

const dateFormat = "2006-01-02"

var (
	proposeTemplate = compileTempl("views/propose.html")
	newUserTemplate = compileTempl("views/new-user.html")
)

type proposePageFiller struct {
	LoggedIn bool
	FullName string
}

// ServeNewBHAPPage serves a page for creating new BHAPs.
func ServeNewBHAPPage(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	// Get the current logged in user
	currUser, userKey, err := bhap.UserFromSession(ctx, r)
	if err != nil {
		http.Error(w, "Could not read session", http.StatusInternalServerError)
		log.Errorf(ctx, "could not get session email: %v", err)
		return
	}

	filler := proposePageFiller{
		LoggedIn: userKey != nil,
		FullName: currUser.FirstName + " " + currUser.LastName,
	}

	showTemplate(ctx, w, proposeTemplate, filler)
}

// HandleNewBHAPForm creates a new BHAP based on information passed from a POST form.
func HandleNewBHAPForm(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	title := r.FormValue("title")
	shortDescription := r.FormValue("shortDescription")
	content := r.FormValue("content")

	// Find what the ID of the new BHAP should be
	newID, err := bhap.NextID(ctx)
	if err != nil {
		log.Errorf(ctx, "could not query BHAPs: %v", err)
		http.Error(w, "Error while finding BHAP", http.StatusInternalServerError)
		return
	}

	// Get the current logged in user
	_, userKey, err := bhap.UserFromSession(ctx, r)
	if err != nil {
		http.Error(w, "Could not read session", http.StatusInternalServerError)
		log.Errorf(ctx, "could not get session email: %v", err)
		return
	}

	newBHAP := bhap.BHAP{
		ID:               newID,
		Title:            title,
		ShortDescription: shortDescription,
		LastModified:     time.Now(),
		Author:           userKey,
		Status:           bhap.DraftStatus,
		CreatedDate:      time.Now(),
		Content:          content,
	}

	// Save the new BHAP
	key := datastore.NewKey(ctx, "BHAP", "", 0, nil)
	if _, err := datastore.Put(ctx, key, &newBHAP); err != nil {
		log.Errorf(ctx, "failed to save BHAP: %v", err)
		http.Error(w, "Could not save BHAP", http.StatusInternalServerError)
		return
	}

	log.Infof(ctx, "Saved BHAP %v: %v", newID, title)

	http.Redirect(w, r, fmt.Sprintf("/bhap/%v", newID), http.StatusSeeOther)
}
