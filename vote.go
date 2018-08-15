package bhap

import (
	"context"
	"fmt"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

const voteEntityName = "Vote"

// Vote represents a user's vote for a BHAP.
type Vote struct {
	OnBHAP *datastore.Key
	ByUser *datastore.Key
	Value  Status
}

// AllVotesForBHAP returns all the votes that have been cast for a given BHAP.
func AllVotesForBHAP(ctx context.Context, bhapKey *datastore.Key) ([]Vote, error) {
	var votes []Vote
	_, err := datastore.NewQuery(voteEntityName).
		Ancestor(bhapKey).
		GetAll(ctx, &votes)
	if err != nil {
		return []Vote{}, fmt.Errorf("getting BHAP votes: %v", err)
	}

	return votes, nil
}

// GetVoteForBHAP returns the user's current vote on a BHAP.
func GetVoteForBHAP(ctx context.Context, bhapKey, userKey *datastore.Key) (Vote, *datastore.Key, error) {
	var votes []Vote
	keys, err := datastore.NewQuery(voteEntityName).
		Ancestor(bhapKey).
		Filter("ByUser =", userKey).
		GetAll(ctx, &votes)
	if err != nil {
		return Vote{}, nil, fmt.Errorf("getting user's vote: %v", err)
	}

	if len(votes) == 0 {
		return Vote{}, nil, nil
	} else {
		return votes[0], keys[0], nil
	}
}

// SetVoteForBHAP sets the vote of the user for the given BHAP to a value,
// creating a new vote object if necessary.
func SetVoteForBHAP(ctx context.Context, bhapKey, userKey *datastore.Key, value Status) error {
	var existingVotes []Vote
	existingKeys, err := datastore.NewQuery(voteEntityName).
		Ancestor(bhapKey).
		Filter("ByUser =", userKey).
		GetAll(ctx, &existingVotes)
	if err != nil {
		return fmt.Errorf("looking for existing votes: %v", err)
	}

	var voteToSave Vote
	var voteKey *datastore.Key

	if len(existingVotes) == 0 {
		// Make a new vote if one doesn't exist
		voteKey = datastore.NewIncompleteKey(ctx, voteEntityName, bhapKey)
		voteToSave = Vote{
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

// CheckVotes counts up all votes for a BHAP and changes its status if
// necessary. All users must vote for the BHAP to be finalized.
func CheckVotes(ctx context.Context, bhapKey *datastore.Key, forBHAP BHAP) error {
	votes, err := AllVotesForBHAP(ctx, bhapKey)
	if err != nil {
		return err
	}

	accepted := 0
	rejected := 0
	for _, vote := range votes {
		if vote.Value == AcceptedStatus {
			accepted++
		} else if vote.Value == RejectedStatus {
			rejected++
		}
	}

	userCnt, err := datastore.NewQuery(UserEntityName).Count(ctx)
	if err != nil {
		return fmt.Errorf("counting users: %v", err)
	}

	nonAuthorCnt := userCnt - 1

	if accepted+rejected == nonAuthorCnt {
		if accepted > nonAuthorCnt/2 {
			forBHAP.Status = AcceptedStatus
			log.Infof(ctx, "marked BHAP %v as accepted", forBHAP.ID)
		} else if rejected > nonAuthorCnt/2 {
			forBHAP.Status = RejectedStatus
			log.Infof(ctx, "marked BHAP %v as rejected", forBHAP.ID)
		}
	}

	if _, err := datastore.Put(ctx, bhapKey, &forBHAP); err != nil {
		return fmt.Errorf("saving BHAP: %v", err)
	}

	return nil
}
