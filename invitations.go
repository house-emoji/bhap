package main

import (
	"context"

	"google.golang.org/appengine/datastore"
)

const InvitationEntityName = "Invitation"

type invitation struct {
	Email     string
	UID       string
	EmailSent bool
}

func unsentInvitations(ctx context.Context) ([]invitation, []*datastore.Key, error) {
	var results []invitation
	query := datastore.NewQuery(InvitationEntityName).
		Filter("EmailSent =", false)

	keys, err := query.GetAll(ctx, &results)
	if err != nil {
		return nil, nil, err
	}

	return results, keys, nil
}

func invitationByUID(ctx context.Context, uid string) (invitation, *datastore.Key, error) {
	var results []invitation
	query := datastore.NewQuery(InvitationEntityName).
		Filter("UID =", uid)

	keys, err := query.GetAll(ctx, &results)
	if err != nil {
		return invitation{}, nil, err
	}

	if len(results) == 0 {
		return invitation{}, nil, nil
	}

	return results[0], keys[0], nil
}
