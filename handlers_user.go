package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/desmondmcnamee/populr_go_api/Godeps/_workspace/src/github.com/gorilla/context"
	"github.com/desmondmcnamee/populr_go_api/Godeps/_workspace/src/github.com/julienschmidt/httprouter"
)

// Returns all users.
func (c *appContext) getUsersHandler(w http.ResponseWriter, r *http.Request) {
	userId := r.Header.Get("x-key")
	var users []ResponseUser
	err := c.db.Select(&users, "SELECT id, username FROM users ORDER BY username ASC")
	if err != nil {
		log.Println("Error finding users: ", err)
		WriteError(w, ErrInternalServer)
	}

	detailedResponseUsers, err := c.MakeDetailResponseUsers(&users, userId)
	if err != nil {
		log.Println("Error searching on users: ", err)
		WriteError(w, ErrInternalServer)
		return
	}

	Respond(w, r, 201, detailedResponseUsers)
}

// Returns a single user.
func (c *appContext) getUserHandler(w http.ResponseWriter, r *http.Request) {
	userId := r.Header.Get("x-key")
	params := context.Get(r, "params").(httprouter.Params)
	var user ResponseUser
	err := c.db.Get(&user, "SELECT id, username FROM users WHERE id=$1", params.ByName("id"))
	if err == sql.ErrNoRows {
		WriteError(w, ErrNoUserForId)
		return
	}
	detailedResponseUsers, err := c.MakeDetailResponseUsers(&[]ResponseUser{user}, userId)
	if err != nil {
		log.Println("Error searching on users: ", err)
		WriteError(w, ErrInternalServer)
		return
	}

	Respond(w, r, 201, detailedResponseUsers)
}

func (c *appContext) loginUserHandler(w http.ResponseWriter, r *http.Request) {
	body := context.Get(r, "body").(*RecieveUserResource)
	user := body.Data

	// Check if this username is already taken.
	var savedUser RecieveUser
	err := c.db.Get(&savedUser, "SELECT * FROM users WHERE username=$1", user.Username)

	// User doesn't exist.
	if err == sql.ErrNoRows {
		WriteError(w, ErrInvalidLogin)
		return
	}

	if err != nil {
		log.Println("Error finding user: ", err)
		WriteError(w, ErrInternalServer)
		return
	}

	// Password is incorrect.
	if savedUser.Password != user.Password {
		WriteError(w, ErrInvalidLogin)
		return
	}

	userToReturn := ResponseUser{User: savedUser.User}

	Respond(w, r, 201, userToReturn)
}

// Creates a user.
func (c *appContext) createUserHandler(w http.ResponseWriter, r *http.Request) {
	body := context.Get(r, "body").(*RecieveUserResource)
	user := body.Data

	// Check if this username is already taken.
	var users []RecieveUser
	c.db.Select(&users, "SELECT * FROM users WHERE username=$1", user.Username)
	if len(users) != 0 {
		WriteError(w, ErrUserExists)
		return
	}

	_, err := c.db.Exec("INSERT INTO users (username, password) VALUES ($1, $2)", user.Username, user.Password)
	if err != nil {
		log.Println("Error creating user: ", err)
		WriteError(w, ErrInternalServer)
		return
	}

	var newUser ResponseUser
	err = c.db.Get(&newUser, "SELECT id, username FROM users WHERE username=$1", user.Username)
	if err != nil {
		log.Println("User created but there was an error on return: ", err)
	}

	Respond(w, r, 201, newUser)
}

// Returns all users that match search
func (c *appContext) searchUsersHandler(w http.ResponseWriter, r *http.Request) {
	userId := r.Header.Get("x-key")
	params := context.Get(r, "params").(httprouter.Params)
	searchTerm := params.ByName("term")
	searchTerm = fmt.Sprint(searchTerm, "%")

	log.Println("Search Term: ", searchTerm)

	var users []ResponseUser
	c.db.Select(&users, "SELECT id, username FROM users WHERE users.username LIKE $1 AND users.id != $2 ORDER BY username ASC", searchTerm, userId)

	detailedResponseUsers, err := c.MakeDetailResponseUsers(&users, userId)
	if err != nil {
		log.Println("Error searching on users: ", err)
		WriteError(w, ErrInternalServer)
		return
	}

	Respond(w, r, 201, detailedResponseUsers)
}
