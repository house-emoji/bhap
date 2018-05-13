package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	blackfriday "gopkg.in/russross/blackfriday.v2"
)

const dateFormat = "2006-01-02"

var (
	listTemplate    = compileTempl("views/list.html")
	bhapTemplate    = compileTempl("views/bhap.html")
	proposeTemplate = compileTempl("views/propose.html")
	newUserTemplate = compileTempl("views/new-user.html")
)

// serveBHAPPage serves up a page that displays info on a single BHAP.
func serveBHAPPage(w http.ResponseWriter, r *http.Request) {
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

	// Render the BHAP content
	html := string(blackfriday.Run([]byte(loadedBHAP.Content)))

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

	// Decide what voting options the user should have
	var votingOpts votingOptions
	if userKey != nil {
		switch loadedBHAP.Status {
		case draftStatus:
			if author.Email == currUser.Email {
				votingOpts.ShowReadyForDiscussion = true
				votingOpts.ShowWithdraw = true
			}
		case discussionStatus:
			if author.Email == currUser.Email {
				votingOpts.ShowWithdraw = true
			} else {
				votingOpts.ShowAccept = true
				votingOpts.ShowReject = true
			}
		case acceptedStatus:
			votingOpts.ShowReplace = true
		}
	}
	log.Infof(ctx, "%+v", votingOpts)

	filler := bhapFiller{
		ID:            loadedBHAP.ID,
		PaddedID:      fmt.Sprintf("%04d", loadedBHAP.ID),
		Title:         loadedBHAP.Title,
		LastModified:  loadedBHAP.LastModified.Format(dateFormat),
		Author:        author.String(),
		Status:        loadedBHAP.Status,
		CreatedDate:   loadedBHAP.CreatedDate.Format(dateFormat),
		VotingOptions: votingOpts,
		HTMLContent:   template.HTML(html),
	}
	showTemplate(ctx, w, bhapTemplate, filler)
}

// serveProposePage serves a page for creating new BHAPs.
func serveProposePage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "views/propose.html")
}

// serveListPage serves a page with a list of all BHAPs.
func serveListPage(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	bhaps, err := allBHAPs(ctx)
	if err != nil {
		log.Errorf(ctx, "could not load BHAPs: %v", err)
		http.Error(w, "Failed to load BHAPs", http.StatusInternalServerError)
		return
	}

	// Get the current logged in user
	currUser, userKey, err := userFromSession(ctx, r)
	if err != nil {
		http.Error(w, "Could not read session", http.StatusInternalServerError)
		log.Errorf(ctx, "could not get session email: %v", err)
		return
	}

	filler := listFiller{
		LoggedIn: userKey != nil,
		Email:    currUser.Email,
		Items:    bhapsToListItemFillers(bhaps),
	}

	showTemplate(ctx, w, listTemplate, filler)
}

// handleNewBHAPForm creates a new BHAP based on information passed from a POST form.
func handleNewBHAPForm(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	title := r.FormValue("title")
	content := r.FormValue("content")

	// Find what the ID of the new BHAP should be
	newID, err := nextID(ctx)
	if err != nil {
		log.Errorf(ctx, "could not query BHAPs: %v", err)
		http.Error(w, "Error while finding BHAP", http.StatusInternalServerError)
		return
	}

	// Get the current logged in user
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

	http.Redirect(w, r, fmt.Sprintf("/bhap/%v", newID), http.StatusSeeOther)
}
