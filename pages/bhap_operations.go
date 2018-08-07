package pages

import (
	"fmt"
	"net/http"

	"github.com/house-emoji/bhap"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

// HandleReadyForDiscussion handles requests to make BHAPs as ready to be
// discussed.
func HandleReadyForDiscussion(op bhapOperator, w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	if !op.bhap.Author.Equal(op.userKey) {
		http.Error(w, "Only authors may mark a BHAP as ready for discussion",
			http.StatusForbidden)
		log.Warningf(ctx, "request from non-author denied")
		return
	}

	if op.bhap.Status != bhap.DraftStatus {
		http.Error(w, "Only drafts may be marked as ready for discussion",
			http.StatusBadRequest)
		log.Warningf(ctx, "non-draft BHAP denied")
		return
	}

	newID, err := bhap.NextID(ctx, op.bhap.Type)
	if err != nil {
		http.Error(w, "Error while assigning new BHAP ID",
			http.StatusInternalServerError)
		log.Errorf(ctx, "getting next BHAP ID: %V", err)
	}

	op.bhap.ID = newID
	op.bhap.Status = bhap.DiscussionStatus
	if _, err := datastore.Put(ctx, op.bhapKey, &op.bhap); err != nil {
		http.Error(w, "Could not update BHAP", http.StatusInternalServerError)
		log.Warningf(ctx, "updating BHAP: %v", err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/bhap/%v", op.bhap.ID), http.StatusSeeOther)
}

// HandleDeleteVote handles requests to delete a submitted vote.
func HandleDeleteVote(op bhapOperator, w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	if op.bhap.Author.Equal(op.userKey) {
		http.Error(w, "Authors may not vote on their own BHAP", http.StatusBadRequest)
		log.Warningf(ctx, "request from author denied")
		return
	}

	if op.bhap.Status != bhap.DiscussionStatus {
		http.Error(w, "Only discussion BHAPs can have votes deleted",
			http.StatusBadRequest)
		log.Warningf(ctx, "vote delete request on non-discussion BHAP denied")
		return
	}

	_, voteKey, err := bhap.GetVoteForBHAP(ctx, op.bhapKey, op.userKey)
	if err != nil {
		http.Error(w, "Could not load vote", http.StatusInternalServerError)
		log.Errorf(ctx, "getting vote: %v", err)
		return
	}
	if voteKey == nil {
		http.Error(w, "No vote has been cast", http.StatusNotFound)
		log.Warningf(ctx, "vote delete request on non-existent vote denied")
		return
	}

	err = datastore.Delete(ctx, voteKey)
	if err != nil {
		http.Error(w, "Could not delete vote", http.StatusInternalServerError)
		log.Errorf(ctx, "deleting vote: %v", err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/bhap/%v", op.bhap.ID), http.StatusSeeOther)
}

// HandleVoteAccept handles requests to submit an accept vote on the BHAP.
func HandleVoteAccept(op bhapOperator, w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	if op.bhap.Author.Equal(op.userKey) {
		http.Error(w, "Authors may not vote on their own BHAP", http.StatusBadRequest)
		log.Warningf(ctx, "request from author denied")
		return
	}

	if op.bhap.Status != bhap.DiscussionStatus {
		http.Error(w, "Only discussion BHAPs may be voted on",
			http.StatusBadRequest)
		log.Warningf(ctx, "vote on non-discussion BHAP denied")
		return
	}

	err := bhap.SetVoteForBHAP(ctx, op.bhapKey, op.userKey, bhap.AcceptedStatus)
	if err != nil {
		log.Errorf(ctx, "could not create vote: %v", err)
		http.Error(w, "Could not create vote", 500)
		return
	}

	bhap.CheckVotes(ctx, op.bhapKey, op.bhap)

	http.Redirect(w, r, fmt.Sprintf("/bhap/%v", op.bhap.ID), http.StatusSeeOther)
}

// HandleVoteReject handles requests to submit an reject vote on the BHAP.
func HandleVoteReject(op bhapOperator, w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	if op.bhap.Author.Equal(op.userKey) {
		http.Error(w, "Authors may not vote on their own BHAP", http.StatusBadRequest)
		log.Warningf(ctx, "request from non-author denied")
		return
	}

	if op.bhap.Status != bhap.DiscussionStatus {
		http.Error(w, "Only discussion BHAPs may be voted on",
			http.StatusBadRequest)
		log.Warningf(ctx, "vote on non-discussion BHAP denied")
		return
	}

	err := bhap.SetVoteForBHAP(ctx, op.bhapKey, op.userKey, bhap.RejectedStatus)
	if err != nil {
		log.Errorf(ctx, "could not create vote: %v", err)
		http.Error(w, "Could not create vote", 500)
		return
	}

	bhap.CheckVotes(ctx, op.bhapKey, op.bhap)

	http.Redirect(w, r, fmt.Sprintf("/bhap/%v", op.bhap.ID), http.StatusSeeOther)
}

// HandleWithdraw handles requests to withdraw a BHAP.
func HandleWithdraw(op bhapOperator, w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	if !op.bhap.Author.Equal(op.userKey) {
		http.Error(w, "Only authors may withdraw a BHAP",
			http.StatusForbidden)
		log.Warningf(ctx, "request from non-author denied")
		return
	}

	if op.bhap.Status != bhap.DiscussionStatus {
		http.Error(w, "Only discussion BHAPs may be withdrawn",
			http.StatusBadRequest)
		log.Warningf(ctx, "vote on non-discussion BHAP denied")
		return
	}

	op.bhap.Status = bhap.WithdrawnStatus
	if _, err := datastore.Put(ctx, op.bhapKey, &op.bhap); err != nil {
		http.Error(w, "Could not update BHAP", http.StatusInternalServerError)
		log.Warningf(ctx, "updating BHAP: %v", err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/bhap/%v", op.bhap.ID), http.StatusSeeOther)
}
