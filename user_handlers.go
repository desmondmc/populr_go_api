package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/desmondmcnamee/populr_go_api/Godeps/_workspace/src/github.com/gorilla/context"
	"github.com/desmondmcnamee/populr_go_api/Godeps/_workspace/src/github.com/julienschmidt/httprouter"
)

// Returns all users.
func (c *appContext) getUsersHandler(w http.ResponseWriter, r *http.Request) {
	var users []User
	err := c.db.Select(&users, "SELECT id, username FROM users ORDER BY id ASC")
	if err != nil {
		log.Println("Error finding users: ", err)
		WriteError(w, ErrInternalServer)
	}
	Respond(w, r, 201, users)
}

// Returns a single user.
func (c *appContext) getUserHandler(w http.ResponseWriter, r *http.Request) {
	params := context.Get(r, "params").(httprouter.Params)
	var user ResponseUser
	c.db.Get(&user, "SELECT id, username FROM users WHERE id=$1", params.ByName("id"))

	Respond(w, r, 201, user)
}

func (c *appContext) loginUserHandler(w http.ResponseWriter, r *http.Request) {
	body := context.Get(r, "body").(*RecieveUserResource)
	user := body.Data

	// Check if this username is already taken.
	var savedUser RecieveUser
	err := c.db.Get(&savedUser, "SELECT * FROM users WHERE username=$1", user.Username)

	if err != nil {
		log.Println("Error finding user: ", err)
		WriteError(w, ErrInternalServer)
		return
	}

	// User doesn't exist.
	if savedUser.Id == 0 {
		WriteError(w, ErrInvalidLogin)
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
	params := context.Get(r, "params").(httprouter.Params)
	searchTerm := params.ByName("term")
	searchTerm = fmt.Sprint(searchTerm, "%")

	log.Println("Search Term: ", searchTerm)

	var users []ResponseUser
	c.db.Select(&users, "SELECT id, username FROM users WHERE users.username LIKE $1", searchTerm)

	Respond(w, r, 201, users)
}
