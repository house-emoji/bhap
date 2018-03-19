package main

import (
	"context"
	"time"

	"google.golang.org/appengine/datastore"
)

type status string

const (
	draftStatus      status = "Draft"
	deferredStatus   status = "Deferred"
	rejectedStatus   status = "Rejected"
	discussionStatus status = "Discussion"
	withdrawnStatus  status = "Withdrawn"
	acceptedStatus   status = "Accepted"
	replacedStatus   status = "Replaced"
	aprilFoolsStatus status = "April Fools"
)

const BHAPEntityName = "BHAP"

type bhap struct {
	ID           int
	Title        string
	LastModified time.Time
	Author       *datastore.Key
	Status       status
	CreatedDate  time.Time
	Content      string
}

// bhapByID gets a BHAP by the given ID unless none exists, in which case
// "exists" equals false.
func bhapByID(ctx context.Context, id int) (output bhap, exists bool, err error) {
	var results []bhap
	query := datastore.NewQuery(BHAPEntityName).
		Filter("ID =", id).
		Limit(1)
	if _, err := query.GetAll(ctx, &results); err != nil {
		return bhap{}, false, err
	}

	if len(results) == 0 {
		return bhap{}, false, nil
	}

	return results[0], true, nil
}

// nextID returns the next unused ID for a new BHAP.
func nextID(ctx context.Context) (int, error) {
	// TODO(velovix): Nasty race condition here. Some kind of database lock
	// should fix this

	var results []bhap
	query := datastore.NewQuery(BHAPEntityName).
		Order("-ID").
		Limit(1)
	if _, err := query.GetAll(ctx, &results); err != nil {
		return 0, err
	}

	if len(results) == 0 {
		return 0, nil
	} else {
		return results[0].ID + 1, nil
	}
}

// allBHAPs returns all BHAPs.
func allBHAPs(ctx context.Context) ([]bhap, error) {
	var results []bhap
	_, err := datastore.NewQuery(BHAPEntityName).
		Order("ID").
		GetAll(ctx, &results)
	if err != nil {
		return []bhap{}, err
	}

	return results, nil
}
