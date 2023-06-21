package handlers

import (
	"net/http"

	"github.com/thomastaylor312/printing-api/store"
	"github.com/thomastaylor312/printing-api/types"
)

type UserHandlers struct {
	db store.DataStore
}

func NewUserHandlers(db store.DataStore) *UserHandlers {
	return &UserHandlers{db: db}
}

// GetUsers gets all users from the database
func (u *UserHandlers) GetUsers(w http.ResponseWriter, r *http.Request) {
	get[*types.User](u.db, "users", w, r)
}

// TODO a get user function that lets the user get their own user info

// AddUser adds a user to the database
func (u *UserHandlers) AddUser(w http.ResponseWriter, r *http.Request) {
	add[*types.User](u.db, "users", w, r, nil, nil)
}

// UpdateUser updates a user in the database
func (u *UserHandlers) UpdateUser(w http.ResponseWriter, r *http.Request) {
	update[*types.User](u.db, "users", w, r, nil, nil)
}

// DeleteUser deletes a user from the database
func (u *UserHandlers) DeleteUser(w http.ResponseWriter, r *http.Request) {
	delete[*types.User](u.db, "users", w, r, nil)
}
