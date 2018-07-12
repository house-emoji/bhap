package main

import (
	"context"
	"fmt"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

const voteEntityName = "Vote"

type vote struct {
	OnBHAP *datastore.Key
	ByUser *datastore.Key
	Value  status
}

// allVotesForBHAP returns all the votes that have been cast for a given BHAP.
func allVotesForBHAP(ctx context.Context, bhapKey *datastore.Key) ([]vote, error) {
	var votes []vote
	_, err := datastore.NewQuery(voteEntityName).
		Ancestor(bhapKey).
		GetAll(ctx, &votes)
	if err != nil {
		return []vote{}, fmt.Errorf("getting BHAP votes: %v", err)
	}

	return votes, nil
}

func voteForBHAP(ctx context.Context, bhapKey, userKey *datastore.Key) (vote, *datastore.Key, error) {
	var votes []vote
	keys, err := datastore.NewQuery(voteEntityName).
		Ancestor(bhapKey).
		Filter("ByUser =", userKey).
		GetAll(ctx, &votes)
	if err != nil {
		return vote{}, nil, fmt.Errorf("getting user's vote: %v", err)
	}

	if len(votes) == 0 {
		return vote{}, nil, nil
	} else {
		return votes[0], keys[0], nil
	}
}

// setVoteForBHAP sets the vote of the user for the given BHAP to a value,
// creating a new vote object if necessary.
func setVoteForBHAP(ctx context.Context, bhapKey, userKey *datastore.Key, value status) error {
	var existingVotes []vote
	existingKeys, err := datastore.NewQuery(voteEntityName).
		Ancestor(bhapKey).
		Filter("ByUser =", userKey).
		GetAll(ctx, &existingVotes)
	if err != nil {
		return fmt.Errorf("looking for existing votes: %v", err)
	}

	var voteToSave vote
	var voteKey *datastore.Key

	if len(existingVotes) == 0 {
		// Make a new vote if one doesn't exist
		voteKey = datastore.NewIncompleteKey(ctx, voteEntityName, bhapKey)
		voteToSave = vote{
			OnBHAP: bhapKey,
			ByUser: userKey,
			Value:  value}
	} else {
		// Edit the existing vote
		voteKey = existingKeys[0]
		voteToSave = existingVotes[0]
		voteToSave.Value = value
	}

	if _, err := datastore.Put(ctx, voteKey, &voteToSave); err != nil {
		return fmt.Errorf("creating vote: %v", err)
	}

	return nil
}

// checkVotes counts up all votes for a BHAP and changes its status if
// necessary. All users must vote for the BHAP to be finalized.
func checkVotes(ctx context.Context, op bhapOperator) error {
	votes, err := allVotesForBHAP(ctx, op.bhapKey)
	if err != nil {
		return err
	}

	accepted := 0
	rejected := 0
	for _, vote := range votes {
		if vote.Value == acceptedStatus {
			accepted++
		} else if vote.Value == rejectedStatus {
			rejected++
		}
	}

	userCnt, err := datastore.NewQuery(userEntityName).Count(ctx)
	if err != nil {
		return fmt.Errorf("counting users: %v", err)
	}

	if accepted+rejected == userCnt-1 {
		if accepted > userCnt/2 {
			op.bhap.Status = acceptedStatus
			log.Infof(ctx, "marked BHAP %v as accepted", op.bhap.ID)
		} else if rejected > userCnt/2 {
			op.bhap.Status = rejectedStatus
			log.Infof(ctx, "marked BHAP %v as rejected", op.bhap.ID)
		}
	}

	if _, err := datastore.Put(ctx, op.bhapKey, &op.bhap); err != nil {
		return fmt.Errorf("saving BHAP: %v", err)
	}

	return nil
}
