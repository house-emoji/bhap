package main

import (
	"fmt"
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

// handleReadyForDiscussion handles requests to make BHAPs as ready to be
// discussed.
func handleReadyForDiscussion(op bhapOperator, w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	if !op.bhap.Author.Equal(op.userKey) {
		http.Error(w, "Only authors may mark a BHAP as ready for discussion",
			http.StatusForbidden)
		log.Warningf(ctx, "request from non-author denied")
		return
	}

	if op.bhap.Status != draftStatus {
		http.Error(w, "Only drafts may be marked as ready for discussion",
			http.StatusBadRequest)
		log.Warningf(ctx, "non-draft BHAP denied")
		return
	}

	op.bhap.Status = discussionStatus
	if _, err := datastore.Put(ctx, op.bhapKey, &op.bhap); err != nil {
		http.Error(w, "Could not update BHAP", http.StatusInternalServerError)
		log.Warningf(ctx, "updating BHAP: %v", err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/bhap/%v", op.bhap.ID), http.StatusSeeOther)
}

// handleDeleteVote handles requests to delete a submitted vote.
func handleDeleteVote(op bhapOperator, w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	if op.bhap.Author.Equal(op.userKey) {
		http.Error(w, "Authors may not vote on their own BHAP", http.StatusBadRequest)
		log.Warningf(ctx, "request from author denied")
		return
	}

	if op.bhap.Status != discussionStatus {
		http.Error(w, "Only discussion BHAPs can have votes deleted",
			http.StatusBadRequest)
		log.Warningf(ctx, "vote delete request on non-discussion BHAP denied")
		return
	}

	_, voteKey, err := voteForBHAP(ctx, op.bhapKey, op.userKey)
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

// handleVoteAccept handles requests to submit an accept vote on the BHAP.
func handleVoteAccept(op bhapOperator, w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	if op.bhap.Author.Equal(op.userKey) {
		http.Error(w, "Authors may not vote on their own BHAP", http.StatusBadRequest)
		log.Warningf(ctx, "request from author denied")
		return
	}

	if op.bhap.Status != discussionStatus {
		http.Error(w, "Only discussion BHAPs may be voted on",
			http.StatusBadRequest)
		log.Warningf(ctx, "vote on non-discussion BHAP denied")
		return
	}

	err := setVoteForBHAP(ctx, op.bhapKey, op.userKey, acceptedStatus)
	if err != nil {
		log.Errorf(ctx, "could not create vote: %v", err)
		http.Error(w, "Could not create vote", 500)
		return
	}

	checkVotes(ctx, op)

	http.Redirect(w, r, fmt.Sprintf("/bhap/%v", op.bhap.ID), http.StatusSeeOther)
}

// handleVoteReject handles requests to submit an reject vote on the BHAP.
func handleVoteReject(op bhapOperator, w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	if op.bhap.Author.Equal(op.userKey) {
		http.Error(w, "Authors may not vote on their own BHAP", http.StatusBadRequest)
		log.Warningf(ctx, "request from non-author denied")
		return
	}

	if op.bhap.Status != discussionStatus {
		http.Error(w, "Only discussion BHAPs may be voted on",
			http.StatusBadRequest)
		log.Warningf(ctx, "vote on non-discussion BHAP denied")
		return
	}

	err := setVoteForBHAP(ctx, op.bhapKey, op.userKey, rejectedStatus)
	if err != nil {
		log.Errorf(ctx, "could not create vote: %v", err)
		http.Error(w, "Could not create vote", 500)
		return
	}

	checkVotes(ctx, op)

	http.Redirect(w, r, fmt.Sprintf("/bhap/%v", op.bhap.ID), http.StatusSeeOther)
}

// handleWithdraw handles requests to withdraw a BHAP.
func handleWithdraw(op bhapOperator, w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	if !op.bhap.Author.Equal(op.userKey) {
		http.Error(w, "Only authors may withdraw a BHAP",
			http.StatusForbidden)
		log.Warningf(ctx, "request from non-author denied")
		return
	}

	if op.bhap.Status != discussionStatus {
		http.Error(w, "Only discussion BHAPs may be withdrawn",
			http.StatusBadRequest)
		log.Warningf(ctx, "vote on non-discussion BHAP denied")
		return
	}

	op.bhap.Status = withdrawnStatus
	if _, err := datastore.Put(ctx, op.bhapKey, &op.bhap); err != nil {
		http.Error(w, "Could not update BHAP", http.StatusInternalServerError)
		log.Warningf(ctx, "updating BHAP: %v", err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/bhap/%v", op.bhap.ID), http.StatusSeeOther)
}
