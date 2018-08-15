package pages

import (
	"net/http"

	"github.com/house-emoji/bhap"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

var listTemplate = compileTempl("views/list.html")

// listPageFiller fills the BHAP list page template.
type listPageFiller struct {
	LoggedIn        bool
	FullName        string
	NewBHAP         *bhap.BHAP
	DiscussionBHAPs []bhap.BHAP
	ActiveBHAPs     []bhap.BHAP
	RejectedBHAPs   []bhap.BHAP
	DraftBHAPs      []bhap.BHAP
}

// ServeListPage serves a page with a list of all BHAPs.
func ServeListPage(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	// Get the current logged in user
	currUser, userKey, err := bhap.UserFromSession(ctx, r)
	if err != nil {
		http.Error(w, "Could not read session", http.StatusInternalServerError)
		log.Errorf(ctx, "could not get session email: %v", err)
		return
	}

	// Get all discussion BHAPs
	discussionBHAPs, err := bhap.ByStatus(ctx, bhap.DiscussionStatus)
	if err != nil {
		log.Errorf(ctx, "getting discussion BHAPs: %v", err)
		http.Error(w, "Could not get discussion BHAPs",
			http.StatusInternalServerError)
		return
	}
	// Feature the first discussion BHAP
	var newBHAP *bhap.BHAP
	if len(discussionBHAPs) > 0 {
		newBHAP = &discussionBHAPs[0]
		discussionBHAPs = discussionBHAPs[1:]
	}

	// Get all active BHAPs
	activeBHAPs, err := bhap.ByStatus(ctx, bhap.AcceptedStatus)
	if err != nil {
		log.Errorf(ctx, "getting active BHAPs: %v", err)
		http.Error(w, "Could not get active BHAPs",
			http.StatusInternalServerError)
		return
	}

	// Get all rejected BHAPs
	rejectedBHAPs, err := bhap.ByStatus(ctx, bhap.RejectedStatus)
	if err != nil {
		log.Errorf(ctx, "getting rejected BHAPs: %v", err)
		http.Error(w, "Could not get rejected BHAPs",
			http.StatusInternalServerError)
		return
	}

	// Get all draft BHAPs
	draftBHAPs, err := bhap.ByStatus(ctx, bhap.DraftStatus)
	if err != nil {
		log.Errorf(ctx, "getting draft BHAPs: %v", err)
		http.Error(w, "Could not get draft BHAPs",
			http.StatusInternalServerError)
		return
	}

	filler := listPageFiller{
		LoggedIn:        userKey != nil,
		FullName:        currUser.FirstName + " " + currUser.LastName,
		NewBHAP:         newBHAP,
		DiscussionBHAPs: discussionBHAPs,
		ActiveBHAPs:     activeBHAPs,
		RejectedBHAPs:   rejectedBHAPs,
		DraftBHAPs:      draftBHAPs,
	}

	showTemplate(ctx, w, listTemplate, filler)
}
