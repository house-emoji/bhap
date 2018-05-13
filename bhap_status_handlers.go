package main

import (
	"context"
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

// handleVoteAccept handles requests to submit an accept vote on the BHAP.
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

	vote := vote{
		onBHAP: op.bhapKey,
		byUser: op.userKey,
		value:  acceptedStatus}
	key := datastore.NewIncompleteKey(ctx, voteEntityName, op.bhapKey)
	if _, err := datastore.Put(ctx, key, &vote); err != nil {
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

	vote := vote{
		onBHAP: op.bhapKey,
		byUser: op.userKey,
		value:  rejectedStatus}
	key := datastore.NewIncompleteKey(ctx, voteEntityName, nil)
	if _, err := datastore.Put(ctx, key, &vote); err != nil {
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

// checkVotes counts up all votes for a BHAP and changes its status if
// necessary.
func checkVotes(ctx context.Context, op bhapOperator) error {
	votes, err := votesForBHAP(ctx, op.bhapKey)
	if err != nil {
		return err
	}

	accepted := 0
	rejected := 0
	for _, vote := range votes {
		if vote.value == acceptedStatus {
			accepted++
		} else if vote.value == rejectedStatus {
			rejected++
		}
	}

	userCnt, err := datastore.NewQuery(userEntityName).Count(ctx)
	if err != nil {
		return fmt.Errorf("counting users: %v", err)
	}

	if accepted > userCnt/2 {
		op.bhap.Status = acceptedStatus
		log.Infof(ctx, "marked BHAP %v as accepted", op.bhap.ID)
	} else if rejected > userCnt/2 {
		op.bhap.Status = rejectedStatus
		log.Infof(ctx, "marked BHAP %v as rejected", op.bhap.ID)
	}

	if _, err := datastore.Put(ctx, op.bhapKey, &op.bhap); err != nil {
		return fmt.Errorf("saving BHAP: %v", err)
	}

	return nil
}
