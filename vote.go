package main

import (
	"context"
	"fmt"

	"google.golang.org/appengine/datastore"
)

const voteEntityName = "Vote"

type vote struct {
	onBHAP *datastore.Key
	byUser *datastore.Key
	value  status
}

func votesForBHAP(ctx context.Context, bhapKey *datastore.Key) ([]vote, error) {
	var votes []vote
	query := datastore.NewQuery(voteEntityName).
		Filter("onBHAP =", bhapKey)
	_, err := query.GetAll(ctx, &votes)
	if err != nil {
		return []vote{}, fmt.Errorf("getting BHAP votes: %v", err)
	}

	return votes, nil
}
