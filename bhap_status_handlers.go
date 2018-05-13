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

	if op.bhap.Author != op.userKey {
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

func handleVoteAccept(op bhapOperator, w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	if op.bhap.Author == op.userKey {
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

	// TODO(velovix): Do more here
}

func handleVoteReject(op bhapOperator, w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	if op.bhap.Author == op.userKey {
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

	// TODO(velovix): Do more here
}

func handleWithdraw(op bhapOperator, w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	if op.bhap.Author != op.userKey {
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
