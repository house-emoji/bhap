package main

import (
	"context"
	"fmt"
	"sort"
	"time"

	"google.golang.org/appengine/datastore"
)

// status describes the current status of a BHAP.
type status string

const (
	// draftStatus is for a BHAP that is still being written.
	draftStatus status = "Draft"
	// deferredStatus is for a BHAP that is on hold.
	deferredStatus status = "Deferred"
	// rejectedStatus is for a BHAP that was rejected during voting.
	rejectedStatus status = "Rejected"
	// discussionStatus is for a BHAP currently being considered.
	discussionStatus status = "Discussion"
	// withdrawnStatus is for a BHAP that was removed by its author.
	withdrawnStatus status = "Withdrawn"
	// acceptedStatus is for a BHAP that was voted on.
	acceptedStatus status = "Accepted"
	// replacedStatus is for a BHAP that was superseded by another BHAP.
	replacedStatus status = "Replaced"
	// aprilFoolsStatus is for a BHAP that should not be taken seriously.
	aprilFoolsStatus status = "April Fools"
)

const bhapEntityName = "BHAP"

// bhap contains info on a BHAP proposal. It is meant to be persisted in
// Datastore.
type bhap struct {
	ID               int
	Title            string
	ShortDescription string
	LastModified     time.Time
	Author           *datastore.Key
	Status           status
	CreatedDate      time.Time
	// Stored in Markdown
	Content string
}

// bhapByID gets a BHAP by the given ID unless none exists, in which case
// "exists" equals false.
func bhapByID(ctx context.Context, id int) (bhap, *datastore.Key, error) {
	var results []bhap
	query := datastore.NewQuery(bhapEntityName).
		Filter("ID =", id).
		Limit(1)
	keys, err := query.GetAll(ctx, &results)
	if err != nil {
		return bhap{}, nil, err
	}

	if len(results) == 0 {
		return bhap{}, nil, nil
	}

	return results[0], keys[0], nil
}

// nextID returns the next unused ID for a new BHAP.
func nextID(ctx context.Context) (int, error) {
	// TODO(velovix): Nasty race condition here. Some kind of database lock
	// should fix this

	var results []bhap
	query := datastore.NewQuery(bhapEntityName).
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

// allBHAPs returns all recorded BHAPs.
func allBHAPs(ctx context.Context) ([]bhap, error) {
	var results []bhap
	_, err := datastore.NewQuery(bhapEntityName).
		Order("ID").
		GetAll(ctx, &results)
	if err != nil {
		return []bhap{}, err
	}

	return results, nil
}

// bhapsByStatus returns all BHAPs with the given status(es).
func bhapsByStatus(ctx context.Context, statuses ...status) ([]bhap, error) {
	allResults := make([][]bhap, len(statuses))

	// Get BHAPs for every status
	for i, status := range statuses {
		_, err := datastore.NewQuery(bhapEntityName).
			Order("ID").
			Filter("Status =", status).
			GetAll(ctx, &allResults[i])
		if err != nil {
			return nil, fmt.Errorf("finding %v BHAPs: %v", status, err)
		}
	}

	// Combine the BHAP collections into a single set
	resultSet := make(map[bhap]bool)
	for _, section := range allResults {
		for _, result := range section {
			resultSet[result] = true
		}
	}

	// Sort the results
	sorted := make([]bhap, len(resultSet))
	for result, _ := range resultSet {
		sorted = append(sorted, result)
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].ID < sorted[j].ID
	})
	return sorted, nil
}
