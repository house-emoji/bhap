package bhap

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"google.golang.org/appengine/datastore"
)

const UserEntityName = "User"

// User contains data on a member of the BHAP consortium.
type User struct {
	FirstName    string
	LastName     string
	Email        string
	PasswordHash []byte
}

func (u User) String() string {
	return fmt.Sprintf("%v %v <%v>", u.FirstName, u.LastName, u.Email)
}

// CheckLogin checks the given login credentials and returns true if they are
// correct.
func CheckLogin(ctx context.Context, email, password string) (bool, error) {
	var results []User
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

// UserByEmail returns the user with the given email. If no user with that
// email exists, the key will be nil.
func UserByEmail(ctx context.Context, email string) (User, *datastore.Key, error) {
	var results []User
	query := datastore.NewQuery(UserEntityName).
		Filter("Email =", email).
		Limit(1)
	keys, err := query.GetAll(ctx, &results)
	if err != nil {
		return User{}, nil, err
	}

	if len(results) == 0 {
		return User{}, nil, nil
	}

	return results[0], keys[0], nil
}
