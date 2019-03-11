package database

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrInvalidUsernameAndPassword when can't login
	ErrInvalidUsernameAndPassword = errors.New("Incorrect username and/or password")
)

// User is the user struct, storing each user
type User struct {
	model
	Email        string `json:"email"`
	FirstName    string `json:"first_name,omitempty"`
	LastName     string `json:"last_name,omitempty"`
	PasswordSalt []byte `json:",omitempty"`
}

// NewActivity returns a new activity
func NewUser(email, password string) (*User, error) {
	createdAt := time.Now()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)

	if err != nil {
		return nil, err
	}

	u := User{
		Email:        email,
		PasswordSalt: hash,
		model:        model{CreatedAt: createdAt, UpdatedAt: createdAt},
	}

	return &u, nil
}

func (d *datastore) AddUser(u *User) (*User, error) {
	_, err := d.Fetch("user", u.Email)

	if err != nil {

		_, err := d.Add("user", u.Email, u)

		// remove the password salt
		u.PasswordSalt = []byte{}

		return u, err
	}

	return nil, errors.New("Email already exists")
}

func (d *datastore) FindUser(email, password string) (*User, error) {
	bytes, err := d.Fetch("user", email)

	if err != nil {
		return nil, ErrInvalidUsernameAndPassword
	}

	var u User
	if json.Unmarshal(bytes, &u) != nil {
		return nil, ErrInvalidUsernameAndPassword
	}

	if err := bcrypt.CompareHashAndPassword(u.PasswordSalt, []byte(password)); err != nil {
		return nil, ErrInvalidUsernameAndPassword
	}

	// remove the password salt
	u.PasswordSalt = []byte{}

	return &u, nil
}

func (u *User) encode() (io.Reader, error) {
	v, err := json.Marshal(u)
	return bytes.NewReader(v), err
}
