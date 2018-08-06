package bhap

import (
	"context"
	"fmt"
	"sort"
	"time"

	"google.golang.org/appengine/datastore"
)

// Status describes the current status of a BHAP.
type Status string

const (
	// draftStatus is for a BHAP that is still being written.
	DraftStatus Status = "Draft"
	// deferredStatus is for a BHAP that is on hold.
	DeferredStatus Status = "Deferred"
	// rejectedStatus is for a BHAP that was rejected during voting.
	RejectedStatus Status = "Rejected"
	// discussionStatus is for a BHAP currently being considered.
	DiscussionStatus Status = "Discussion"
	// withdrawnStatus is for a BHAP that was removed by its author.
	WithdrawnStatus Status = "Withdrawn"
	// acceptedStatus is for a BHAP that was voted on.
	AcceptedStatus Status = "Accepted"
	// replacedStatus is for a BHAP that was superseded by another BHAP.
	ReplacedStatus Status = "Replaced"
	// aprilFoolsStatus is for a BHAP that should not be taken seriously.
	AprilFoolsStatus Status = "April Fools"
)

// BHAPType describes the type of a BHAP.
type BHAPType string

const (
	// MetaBHAPType describes a BHAP that is used to describe the BHAP process
	// itself.
	MetaBHAPType = "Meta"
	// HouseRuleBHAPTYpe describes a BHAP that creates a rule that house
	// members must follow. Most BHAPs will be of this type.
	HouseRuleBHAPType = "House Rule"
)

const BHAPEntityName = "BHAP"

// BHAP contains info on a BHAP proposal. It is meant to be persisted in
// Datastore.
type BHAP struct {
	// DraftID is the ID to refer to this BHAP by before it leaves the draft
	// stage and is assigned a normal ID
	DraftID string
	// ID is the primary identifier for BHAPs that are not in the draft stage
	ID               int
	Title            string
	ShortDescription string
	LastModified     time.Time
	Author           *datastore.Key
	Status           Status
	CreatedDate      time.Time
	Type             BHAPType
	// Stored in Markdown
	Content string `datastore:"Content,noindex"`
}

// ByDraftID gets a BHAP by the given draft ID. If none exists, the key will
// equal nil.
func ByDraftID(ctx context.Context, draftID string) (BHAP, *datastore.Key, error) {
	var results []BHAP
	query := datastore.NewQuery(BHAPEntityName).
		Filter("DraftID =", draftID).
		Limit(1)
	keys, err := query.GetAll(ctx, &results)
	if err != nil {
		return BHAP{}, nil, fmt.Errorf("by draft ID %v: %v", draftID, err)
	}

	if len(results) == 0 {
		return BHAP{}, nil, nil
	}

	return results[0], keys[0], nil
}

// ByID gets a BHAP by the given ID. If none exists, the key will equal nil.
func ByID(ctx context.Context, id int) (BHAP, *datastore.Key, error) {
	var results []BHAP
	query := datastore.NewQuery(BHAPEntityName).
		Filter("ID =", id).
		Limit(1)
	keys, err := query.GetAll(ctx, &results)
	if err != nil {
		return BHAP{}, nil, fmt.Errorf("by ID %v: %v", id, err)
	}

	if len(results) == 0 {
		return BHAP{}, nil, nil
	}

	return results[0], keys[0], nil
}

// NextID returns the next unused ID for a new BHAP.
func NextID(ctx context.Context) (int, error) {
	// TODO(velovix): Nasty race condition here. Some kind of database lock
	// should fix this

	var results []BHAP
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

// GetAll returns all recorded BHAPs.
func GetAll(ctx context.Context) ([]BHAP, error) {
	var results []BHAP
	_, err := datastore.NewQuery(BHAPEntityName).
		Order("ID").
		GetAll(ctx, &results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// ByStatus returns all BHAPs with the given status(es).
func ByStatus(ctx context.Context, statuses ...Status) ([]BHAP, error) {
	allResults := make([][]BHAP, len(statuses))

	// Get BHAPs for every status
	for i, status := range statuses {
		_, err := datastore.NewQuery(BHAPEntityName).
			Order("ID").
			Filter("Status =", status).
			GetAll(ctx, &allResults[i])
		if err != nil {
			return nil, fmt.Errorf("finding %v BHAPs: %v", status, err)
		}
	}

	// Combine the BHAP collections into a single set
	resultSet := make(map[BHAP]bool)
	for _, section := range allResults {
		for _, result := range section {
			resultSet[result] = true
		}
	}

	// Sort the results
	sorted := make([]BHAP, 0)
	for result, _ := range resultSet {
		sorted = append(sorted, result)
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].ID < sorted[j].ID
	})

	return sorted, nil
}
