package main

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"google.golang.org/appengine/datastore"
)

const UserEntityName = "User"

type user struct {
	FirstName    string
	LastName     string
	Email        string
	PasswordHash []byte
}

func (u user) String() string {
	return fmt.Sprintf("%v %v <%v>", u.FirstName, u.LastName, u.Email)
}

// checkLogin checks the given login credentials and returns true if they are
// correct.
func checkLogin(ctx context.Context, email, password string) (bool, error) {
	var results []user
	query := datastore.NewQuery(UserEntityName).
		Filter("Email =", email).
		Limit(1)
	if _, err := query.GetAll(ctx, &results); err != nil {
		return false, err
	}

	if len(results) == 0 {
		return false, nil
	}

	err := bcrypt.CompareHashAndPassword(results[0].PasswordHash, []byte(password))

	return err == nil, nil
}

func userByEmail(ctx context.Context, email string) (user, *datastore.Key, error) {
	var results []user
	query := datastore.NewQuery(UserEntityName).
		Filter("Email =", email).
		Limit(1)
	keys, err := query.GetAll(ctx, &results)
	if err != nil {
		return user{}, nil, err
	}

	if len(results) == 0 {
		return user{}, nil, nil
	}

	return results[0], keys[0], nil
}
