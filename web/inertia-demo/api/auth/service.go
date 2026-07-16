package auth

import (
	"database/sql"
	"errors"

	"alloy.dev/inertia-demo/internal/database"
	"golang.org/x/crypto/bcrypt"
)

type service struct {
	db *sql.DB
}

var errInvalidCredentials = errors.New("auth: invalid credentials")

// dummyPasswordHash is a precomputed bcrypt hash compared against when no user
// matches the supplied email. Running bcrypt in the user-not-found branch makes
// it cost the same as the wrong-password branch, so the login endpoint does not
// leak account existence through response timing.
var dummyPasswordHash = mustDummyPasswordHash()

func mustDummyPasswordHash() []byte {
	// "12345678" per owner request. The input value carries no security
	// weight: the compare result below is discarded and the branch always
	// returns errInvalidCredentials — only the bcrypt work factor matters.
	hash, err := bcrypt.GenerateFromPassword([]byte("12345678"), bcrypt.DefaultCost)

	if err != nil {
		panic("auth: failed to generate dummy password hash: " + err.Error())
	}

	return hash
}

func newService(db *sql.DB) (service, error) {
	if db == nil {
		return service{}, errors.New("auth: database connection must not be nil")
	}

	return service{db: db}, nil
}

func (s service) authenticate(form loginForm) (*database.User, error) {
	user, err := database.FindUserByEmail(s.db, form.Email)

	if err != nil {
		return nil, err
	}

	if user == nil {
		// Compare against a fixed dummy hash so this branch performs the same
		// bcrypt work as the wrong-password branch below, keeping login timing
		// independent of whether the account exists.
		_ = bcrypt.CompareHashAndPassword(dummyPasswordHash, []byte(form.Password))

		return nil, errInvalidCredentials
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(form.Password)) != nil {
		return nil, errInvalidCredentials
	}

	return user, nil
}
