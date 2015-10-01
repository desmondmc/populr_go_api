package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/desmondmcnamee/populr_go_api/Godeps/_workspace/src/golang.org/x/crypto/bcrypt"

	"github.com/desmondmcnamee/populr_go_api/Godeps/_workspace/src/github.com/gorilla/context"
	"github.com/desmondmcnamee/populr_go_api/Godeps/_workspace/src/github.com/julienschmidt/httprouter"
)

const PopulrUserId = "34"

// Returns all users.
func (c *appContext) getUsersHandler(w http.ResponseWriter, r *http.Request) {
	userId := r.Header.Get("x-key")

	users, err := c.getAllUsers(userId)
	if err != nil {
		log.Println("Error getting all users: ", err)
		WriteError(w, ErrInternalServer)
		return
	}

	Respond(w, r, 201, users)
}

// Returns a single user.
func (c *appContext) getUserHandler(w http.ResponseWriter, r *http.Request) {
	userId := r.Header.Get("x-key")

	user, err := c.getUserWithId(userId)
	if err != nil {
		log.Println("Error searching on users: ", err)
		WriteError(w, ErrInternalServer)
		return
	}

	Respond(w, r, 201, *user)
}

func (c *appContext) loginUserHandler(w http.ResponseWriter, r *http.Request) {
	body := context.Get(r, "body").(*RecieveUserResource)
	user := body.Data

	// Check if this username exists
	savedUser, err := c.getUserWithUsername(user.Username)
	if err == sql.ErrNoRows {
		// User doesn't exist.
		log.Println("Invalid login attempt.")
		WriteError(w, ErrInvalidLogin)
		return
	}
	if err != nil {
		log.Println("Error finding user: ", err)
		WriteError(w, ErrInternalServer)
		return
	}

	// Validate password.
	err = bcrypt.CompareHashAndPassword([]byte(savedUser.Password), []byte(user.Password))
	if err != nil {
		// Password is incorrect.
		log.Println("Error on login: ", err)
		WriteError(w, ErrInvalidLogin)
		return
	}

	// Get user with phone and token
	userToReturn, err := c.getPhoneTokenUserWithUsername(user.Username)
	if userToReturn == nil {
		log.Println("Failed to retrieve phone user: ", err)
		WriteError(w, ErrInvalidLogin)
		return
	}

	Respond(w, r, 201, *userToReturn)
}

// Creates a user.
func (c *appContext) createUserHandler(w http.ResponseWriter, r *http.Request) {
	body := context.Get(r, "body").(*RecieveUserResource)
	user := body.Data

	// Check if this username is already taken.
	_, err := c.getUserWithUsername(user.Username)
	if err != sql.ErrNoRows {
		log.Println("Signup failed ", err)
		WriteError(w, ErrUserExists)
		return
	}

	// Create the user
	err = c.createUser(user.Username, user.Password)
	if err != nil {
		log.Println("Error creating user: ", err)
		WriteError(w, ErrInternalServer)
		return
	}

	newUser, err := c.getPhoneTokenUserWithUsername(user.Username)
	if err != nil {
		log.Println("User created but there was an error on return: ", err)
	}

	newUserId := fmt.Sprintf("%d", newUser.Id)
	c.addDefaultMessages(newUserId)

	Respond(w, r, 201, *newUser)
}

// Returns all users that match search
func (c *appContext) searchUsersHandler(w http.ResponseWriter, r *http.Request) {
	userId := r.Header.Get("x-key")
	params := context.Get(r, "params").(httprouter.Params)
	searchTerm := params.ByName("term")
	searchTerm = fmt.Sprint(searchTerm, "%")

	log.Println("Search Term: ", searchTerm)

	var users []User
	c.db.Select(
		&users,
		"SELECT id, username FROM users WHERE users.username LIKE $1 AND users.id != $2 ORDER BY username ASC",
		searchTerm,
		userId,
	)

	detailedResponseUsers, err := c.makeDetailResponseUsers(&users, userId)
	if err != nil {
		log.Println("Error searching on users: ", err)
		WriteError(w, ErrInternalServer)
		return
	}

	Respond(w, r, 201, detailedResponseUsers)
}

func (c *appContext) postFeedbackHandler(w http.ResponseWriter, r *http.Request) {
	userId := r.Header.Get("x-key")
	body := context.Get(r, "body").(*RecieveFeedbackResource)
	feedback := body.Data

	_, err := c.db.Exec("INSERT INTO feedbacks (user_id, feedback) VALUES ($1, $2)", userId, feedback.Feedback)
	if err != nil {
		log.Println("Error: ", err)
		WriteError(w, ErrInternalServer)
		return
	}

	Respond(w, r, 204, nil)
}

func (c *appContext) postDeviceTokenHandler(w http.ResponseWriter, r *http.Request) {
	params := context.Get(r, "params").(httprouter.Params)
	userId := r.Header.Get("x-key")
	deviceToken := params.ByName("token")

	_, err := c.db.Exec("UPDATE users SET device_token = $1 WHERE id = $2", deviceToken, userId)
	if err != nil {
		log.Println("Error setting device token: ", err)
		WriteError(w, ErrInternalServer)
		return
	}

	Respond(w, r, 204, nil)
}

func (c *appContext) postPhoneNumberHandler(w http.ResponseWriter, r *http.Request) {
	userId := r.Header.Get("x-key")
	body := context.Get(r, "body").(*RecievePhoneNumberResource)
	phoneNumber := body.Data

	_, err := c.db.Exec(
		"UPDATE users SET phone_number = $1 WHERE id = $2",
		phoneNumber.PhoneNumber,
		userId,
	)

	if err != nil {
		log.Println("Error setting phone number: ", err)
		WriteError(w, ErrInternalServer)
		return
	}

	Respond(w, r, 204, nil)
}

func (c *appContext) postContactsHandler(w http.ResponseWriter, r *http.Request) {
	userId := r.Header.Get("x-key")
	body := context.Get(r, "body").(*RecieveContacts)
	contacts := body.Data

	response, err := c.processContacts(contacts, userId)
	if err != nil {
		log.Println("Error setting device token: ", err)
		WriteError(w, ErrInternalServer)
		return
	}

	Respond(w, r, 201, response)
}

func (c *appContext) logoutHandler(w http.ResponseWriter, r *http.Request) {
	userId := r.Header.Get("x-key")

	_, err := c.db.Exec("UPDATE users SET device_token = $1 WHERE id = $2", "", userId)
	if err != nil {
		log.Println("Error setting device token: ", err)
		WriteError(w, ErrInternalServer)
		return
	}

	Respond(w, r, 204, nil)
}
