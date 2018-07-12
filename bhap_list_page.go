package main

import (
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

var listTemplate = compileTempl("views/list.html")

// listPageFiller fills the BHAP list page template.
type listPageFiller struct {
	LoggedIn      bool
	FullName      string
	NewBHAP       *bhap
	ActiveBHAPs   []bhap
	RejectedBHAPs []bhap
	DraftBHAPs    []bhap
}

// serveListPage serves a page with a list of all BHAPs.
func serveListPage(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	// Get the current logged in user
	currUser, userKey, err := userFromSession(ctx, r)
	if err != nil {
		http.Error(w, "Could not read session", http.StatusInternalServerError)
		log.Errorf(ctx, "could not get session email: %v", err)
		return
	}

	discussionBHAPs, err := bhapsByStatus(ctx, discussionStatus)
	if err != nil {
		log.Errorf(ctx, "getting discussion BHAPs: %v", err)
		http.Error(w, "Could not get discussion BHAPs",
			http.StatusInternalServerError)
		return
	}
	var newBHAP *bhap
	if len(discussionBHAPs) > 0 {
		newBHAP = &discussionBHAPs[0]
	}

	activeBHAPs, err := bhapsByStatus(ctx, acceptedStatus)
	if err != nil {
		log.Errorf(ctx, "getting active BHAPs: %v", err)
		http.Error(w, "Could not get active BHAPs",
			http.StatusInternalServerError)
		return
	}
	rejectedBHAPs, err := bhapsByStatus(ctx, rejectedStatus)
	if err != nil {
		log.Errorf(ctx, "getting rejected BHAPs: %v", err)
		http.Error(w, "Could not get rejected BHAPs",
			http.StatusInternalServerError)
		return
	}
	draftBHAPs, err := bhapsByStatus(ctx, draftStatus)
	if err != nil {
		log.Errorf(ctx, "getting draft BHAPs: %v", err)
		http.Error(w, "Could not get draft BHAPs",
			http.StatusInternalServerError)
		return
	}

	log.Infof(ctx, "Drafts: %+v", draftBHAPs)

	filler := listPageFiller{
		LoggedIn:      userKey != nil,
		FullName:      currUser.FirstName + " " + currUser.LastName,
		NewBHAP:       newBHAP,
		ActiveBHAPs:   activeBHAPs,
		RejectedBHAPs: rejectedBHAPs,
		DraftBHAPs:    draftBHAPs,
	}

	showTemplate(ctx, w, listTemplate, filler)
}
