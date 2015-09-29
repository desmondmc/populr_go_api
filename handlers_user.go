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

const PopulrUserId = "3"
const DefaultMessage1Id = "1"
const DefaultMessage2Id = "2"

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
	err := c.db.Get(&savedUser, "SELECT id, username, password FROM users WHERE username=$1", user.Username)

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

	err = bcrypt.CompareHashAndPassword([]byte(savedUser.Password), []byte(user.Password))
	if err != nil {
		// Password is incorrect.
		WriteError(w, ErrInvalidLogin)
		return
	}

	var userToReturn PhoneUser
	err = c.db.Get(&userToReturn, "SELECT id, username, phone_number FROM users WHERE username=$1", user.Username)
	if err != nil {
		log.Println("Failed to retrieve phone user.")
		WriteError(w, ErrInvalidLogin)
		return
	}

	Respond(w, r, 201, userToReturn)
}

// Creates a user.
func (c *appContext) createUserHandler(w http.ResponseWriter, r *http.Request) {
	body := context.Get(r, "body").(*RecieveUserResource)
	user := body.Data

	// Check if this username is already taken.
	var users []RecieveUser
	c.db.Select(&users, "SELECT id, username, password FROM users WHERE username=$1", user.Username)
	if len(users) != 0 {
		WriteError(w, ErrUserExists)
		return
	}

	// Generate Hash From Password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Error hashing user password: ", err)
		WriteError(w, ErrInternalServer)
		return
	}

	// Create the user
	_, err = c.db.Exec(
		"INSERT INTO users (username, password) VALUES ($1, $2)",
		user.Username,
		string(hashedPassword),
	)
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

	newUserId := fmt.Sprintf("%d", newUser.Id)
	c.addDefaultMessages(newUserId)

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

func (c *appContext) addDefaultMessages(toUserId string) {
	var message1 Message
	message1.Type = "direct"
	message1.Message = "Welcome to POPULR! the fastest messaging app in the world! ✋🏽 ✊🏿 👌🏻 ✌🏾 👍🏿 👊🏽 "

	var message1Id int
	c.db.Get(
		&message1Id,
		"INSERT INTO messages (from_user_id, message, type) VALUES ($1, $2, $3) RETURNING id",
		PopulrUserId,
		message1.Message,
		message1.Type,
	)
	c.db.Exec(
		"INSERT INTO message_to_users (user_id, message_id) VALUES ($1, $2)",
		toUserId,
		message1Id,
	)
}
