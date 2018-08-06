package pages

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/house-emoji/bhap"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	blackfriday "gopkg.in/russross/blackfriday.v2"
)

var bhapTemplate = compileTempl("views/bhap.html")

type optionsMode string

const (
	modeNotLoggedIn      optionsMode = "notLoggedIn"
	modeDraftNotAuthor               = "draftNotAuthor"
	modeDraftAuthor                  = "draftAuthor"
	modeDiscussionAuthor             = "discussionAuthor"
	modeDisucssionNoVote             = "discussionNoVote"
	modeDiscussionVoted              = "discussionVoted"
	modeAccepted                     = "accepted"
	modeRejected                     = "rejected"
)

// bhapPageFiller fills the BHAP viewer page template.
type bhapPageFiller struct {
	LoggedIn     bool
	FullName     string
	ID           int
	BHAP         bhap.BHAP
	SelectedVote string
	OptionsMode  optionsMode
	Editable     bool
	HTMLContent  template.HTML

	VoteCount int
	UserCount int

	PercentAccepted  int
	PercentRejected  int
	PercentUndecided int
}

// ServeBHAPPage serves up a page that displays info on a single BHAP as a
// draft.
func ServeDraftBHAPPage(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	// Get the requested draftID
	draftID := mux.Vars(r)["draftID"]

	loadedBHAP, bhapKey, err := bhap.ByDraftID(ctx, draftID)
	if err != nil {
		log.Errorf(ctx, "could not load BHAP: %v", err)
		http.Error(w, "Failed to load BHAP", http.StatusInternalServerError)
		return
	}
	if bhapKey == nil {
		http.Error(w, fmt.Sprintf("No BHAP with draft ID %v", draftID),
			http.StatusNotFound)
		log.Warningf(ctx, "unknown draft BHAP requested: %v", draftID)
		return
	}

	// Render the BHAP content
	options := blackfriday.WithExtensions(blackfriday.HardLineBreak)
	html := string(blackfriday.Run([]byte(loadedBHAP.Content), options))

	var author bhap.User
	if err := datastore.Get(ctx, loadedBHAP.Author, &author); err != nil {
		log.Errorf(ctx, "loading user: %v", err)
		http.Error(w, "Failed to load user", http.StatusInternalServerError)
		return
	}

	// Get the current logged in user
	user, userKey, err := bhap.UserFromSession(ctx, r)
	if err != nil {
		http.Error(w, "Could not read session", http.StatusInternalServerError)
		log.Errorf(ctx, "getting session email: %v", err)
		return
	}

	var fullName string
	if userKey != nil {
		fullName = user.FirstName + " " + user.LastName
	}

	var mode optionsMode
	if userKey == nil {
		mode = modeNotLoggedIn
	} else if userKey.Equal(loadedBHAP.Author) {
		mode = modeDraftAuthor
	} else {
		mode = modeDraftNotAuthor
	}

	editable := isEditableStatus(loadedBHAP.Status) && userKey.Equal(loadedBHAP.Author)

	filler := bhapPageFiller{
		LoggedIn:    userKey != nil,
		FullName:    fullName,
		ID:          loadedBHAP.ID,
		BHAP:        loadedBHAP,
		OptionsMode: mode,
		Editable:    editable,
		HTMLContent: template.HTML(html),
	}
	showTemplate(ctx, w, bhapTemplate, filler)
}

// ServeBHAPPage serves up a page that displays info on a single BHAP.
func ServeBHAPPage(w http.ResponseWriter, r *http.Request) {
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
	loadedBHAP, bhapKey, err := bhap.ByID(ctx, id)
	if err != nil {
		log.Errorf(ctx, "could not load BHAP: %v", err)
		http.Error(w, "Failed to load BHAP", http.StatusInternalServerError)
		return
	}
	if bhapKey == nil {
		http.Error(w, fmt.Sprintf("No BHAP with ID %v", id),
			http.StatusNotFound)
		log.Warningf(ctx, "unknown BHAP requested: %v", id)
		return
	}

	// Render the BHAP content
	options := blackfriday.WithExtensions(blackfriday.HardLineBreak)
	html := string(blackfriday.Run([]byte(loadedBHAP.Content), options))

	var author bhap.User
	if err := datastore.Get(ctx, loadedBHAP.Author, &author); err != nil {
		log.Errorf(ctx, "loading user: %v", err)
		http.Error(w, "Failed to load user", http.StatusInternalServerError)
		return
	}

	// Get the current logged in user
	user, userKey, err := bhap.UserFromSession(ctx, r)
	if err != nil {
		http.Error(w, "Could not read session", http.StatusInternalServerError)
		log.Errorf(ctx, "getting session email: %v", err)
		return
	}

	allVotes, err := bhap.AllVotesForBHAP(ctx, bhapKey)
	if err != nil {
		http.Error(w, "Could not get votes",
			http.StatusInternalServerError)
		log.Errorf(ctx, "getting votes: %v", err)
		return
	}

	userCount, err := datastore.NewQuery(bhap.UserEntityName).Count(ctx)
	if err != nil {
		http.Error(w, "Could not get user count",
			http.StatusInternalServerError)
		log.Errorf(ctx, "getting user count: %v", err)
		return
	}

	usersVote, usersVoteKey, err := bhap.GetVoteForBHAP(ctx, bhapKey, userKey)
	if err != nil {
		http.Error(w, "Could not read user's vote",
			http.StatusInternalServerError)
		log.Errorf(ctx, "getting user's vote: %v", err)
		return
	}

	// Decide what options the user should have
	var mode optionsMode
	if userKey == nil {
		mode = modeNotLoggedIn
	} else if loadedBHAP.Status == bhap.DraftStatus {
		if userKey.Equal(loadedBHAP.Author) {
			mode = modeDraftAuthor
		} else {
			mode = modeDraftNotAuthor
		}
	} else if loadedBHAP.Status == bhap.DiscussionStatus {
		if userKey.Equal(loadedBHAP.Author) {
			mode = modeDiscussionAuthor
		} else {
			if usersVoteKey == nil {
				mode = modeDisucssionNoVote
			} else {
				mode = modeDiscussionVoted
			}
		}
	} else if loadedBHAP.Status == bhap.AcceptedStatus {
		mode = modeAccepted
	} else if loadedBHAP.Status == bhap.RejectedStatus {
		mode = modeRejected
	}

	// Figure out the vote breakdown
	acceptedCount := 0
	rejectedCount := 0
	for _, v := range allVotes {
		if v.Value == bhap.AcceptedStatus {
			acceptedCount++
		} else if v.Value == bhap.RejectedStatus {
			rejectedCount++
		}
	}
	undecidedCount := userCount - (acceptedCount + rejectedCount) - 1

	var fullName string
	if userKey != nil {
		fullName = user.FirstName + " " + user.LastName
	}

	var selectedVote string
	if usersVoteKey != nil {
		if usersVote.Value == bhap.AcceptedStatus {
			selectedVote = "ACCEPT"
		} else if usersVote.Value == bhap.RejectedStatus {
			selectedVote = "REJECTED"
		} else {
			http.Error(w, "Unknown vote type",
				http.StatusInternalServerError)
			log.Errorf(ctx, "unknown vote type %v", usersVote.Value)
			return
		}
	}

	var percentAccepted, percentRejected, percentUndecided int
	countBesidesAuthor := float64(userCount - 1)
	if countBesidesAuthor != 0 {
		percentAccepted = int((float64(acceptedCount) / countBesidesAuthor) * 100)
		percentRejected = int((float64(rejectedCount) / countBesidesAuthor) * 100)
		percentUndecided = int((float64(undecidedCount) / countBesidesAuthor) * 100)
	}

	editable := isEditableStatus(loadedBHAP.Status) && userKey.Equal(loadedBHAP.Author)

	filler := bhapPageFiller{
		LoggedIn:     userKey != nil,
		FullName:     fullName,
		ID:           loadedBHAP.ID,
		BHAP:         loadedBHAP,
		OptionsMode:  mode,
		SelectedVote: selectedVote,
		Editable:     editable,
		HTMLContent:  template.HTML(html),

		VoteCount: len(allVotes),
		UserCount: userCount - 1,

		PercentAccepted:  percentAccepted,
		PercentRejected:  percentRejected,
		PercentUndecided: percentUndecided,
	}
	showTemplate(ctx, w, bhapTemplate, filler)
}
