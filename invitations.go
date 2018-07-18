package bhap

import (
	"context"

	"google.golang.org/appengine/datastore"
)

const InvitationEntityName = "Invitation"

type Invitation struct {
	Email     string
	UID       string
	EmailSent bool
}

// UnsentInvitations returns all invitations that have yet to be emailed.
func UnsentInvitations(ctx context.Context) ([]Invitation, []*datastore.Key, error) {
	var results []Invitation
	query := datastore.NewQuery(InvitationEntityName).
		Filter("EmailSent =", false)

	keys, err := query.GetAll(ctx, &results)
	if err != nil {
		return nil, nil, err
	}

	return results, keys, nil
}

// InvitationByUID returns the invitation with the corresponding UID.
func InvitationByUID(ctx context.Context, uid string) (Invitation, *datastore.Key, error) {
	var results []Invitation
	query := datastore.NewQuery(InvitationEntityName).
		Filter("UID =", uid)

	keys, err := query.GetAll(ctx, &results)
	if err != nil {
		return Invitation{}, nil, err
	}

	if len(results) == 0 {
		return Invitation{}, nil, nil
	}

	return results[0], keys[0], nil
}
