package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/context"
	"github.com/julienschmidt/httprouter"
)

// Returns all users.
func (c *appContext) getUsersHandler(w http.ResponseWriter, r *http.Request) {
	var users []User
	c.db.Select(&users, "SELECT * FROM users ORDER BY id ASC")

	Respond(w, r, 201, UsersResource{Users: users})
}

// Returns a single user.
func (c *appContext) getUserHandler(w http.ResponseWriter, r *http.Request) {
	params := context.Get(r, "params").(httprouter.Params)
	var user User
	c.db.Get(&user, "SELECT * FROM users WHERE id=$1", params.ByName("id"))

	Respond(w, r, 201, user)
}

// Creates a user.
func (c *appContext) createUserHandler(w http.ResponseWriter, r *http.Request) {
	body := context.Get(r, "body").(*UserResource)
	user := body.Data

	// Check if this username is already taken.
	users := []User{}
	c.db.Select(&users, "SELECT * FROM users WHERE username=$1", user.Username)
	if len(users) != 0 {
		WriteError(w, ErrUserExists)
		return
	}

	tx := c.db.MustBegin()
	tx.MustExec("INSERT INTO users (username, password) VALUES ($1, $2)", user.Username, user.Password)
	err := tx.Commit()
	if err != nil {
		log.Println("Error creating user: ", err)
		WriteError(w, ErrInternalServer)
		return
	}

	var newUser User
	c.db.Get(&newUser, "SELECT * FROM users WHERE username=$1", user.Username)

	Respond(w, r, 201, newUser)
}

// Returns all users that match search
func (c *appContext) searchUsersHandler(w http.ResponseWriter, r *http.Request) {
	params := context.Get(r, "params").(httprouter.Params)
	searchTerm := params.ByName("term")
	searchTerm = fmt.Sprint(searchTerm, "%")

	log.Println("Search Term: ", searchTerm)

	var users []User
	c.db.Select(&users, "SELECT * FROM users WHERE users.username LIKE $1", searchTerm)

	Respond(w, r, 201, users)
}