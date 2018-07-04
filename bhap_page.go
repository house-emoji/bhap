package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	blackfriday "gopkg.in/russross/blackfriday.v2"
)

var bhapTemplate = compileTempl("views/bhap.html")

// bhapPageFiller fills the BHAP viewer page template.
type bhapPageFiller struct {
	ID            int
	PaddedID      string
	Title         string
	LastModified  string
	Author        string
	Status        status
	CreatedDate   string
	VotingOptions votingOptions
	Editable      bool
	HTMLContent   template.HTML
}

// votingOptions configures what voting options the user has in the BHAP
// screen.
type votingOptions struct {
	ShowReadyForDiscussion bool
	ShowAccept             bool
	ShowReject             bool
	ShowWithdraw           bool
	ShowReplace            bool
}

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
	options := blackfriday.WithExtensions(blackfriday.HardLineBreak)
	html := string(blackfriday.Run([]byte(loadedBHAP.Content), options))

	var author user
	if err := datastore.Get(ctx, loadedBHAP.Author, &author); err != nil {
		log.Errorf(ctx, "Error loading user: %v", err)
		http.Error(w, "Failed to load user", http.StatusInternalServerError)
		return
	}

	// Get the current logged in user
	_, userKey, err := userFromSession(ctx, r)
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
			if loadedBHAP.Author.Equal(userKey) {
				votingOpts.ShowReadyForDiscussion = true
				votingOpts.ShowWithdraw = true
			}
		case discussionStatus:
			if loadedBHAP.Author.Equal(userKey) {
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

	filler := bhapPageFiller{
		ID:            loadedBHAP.ID,
		PaddedID:      fmt.Sprintf("%04d", loadedBHAP.ID),
		Title:         loadedBHAP.Title,
		LastModified:  loadedBHAP.LastModified.Format(dateFormat),
		Author:        author.String(),
		Status:        loadedBHAP.Status,
		CreatedDate:   loadedBHAP.CreatedDate.Format(dateFormat),
		VotingOptions: votingOpts,
		Editable:      isEditable(loadedBHAP.Status),
		HTMLContent:   template.HTML(html),
	}
	showTemplate(ctx, w, bhapTemplate, filler)
}
