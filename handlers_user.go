package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"regexp"

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

	if len(user.Password) < 4 {
		log.Println("Signup failed password too short")
		WriteError(w, ErrShortPassword)
		return
	}

	// Letters and numbers only. 3-16 characters.
	match, _ := regexp.MatchString("^[a-z0-9]{3,16}$", user.Username)
	if match != true {
		log.Println("Signup failed invalid username.")
		WriteError(w, ErrInvalidUsername)
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

	// Increment Register Count
	UserCount.Inc()
}

// Returns all users that match search
func (c *appContext) searchUsersHandler(w http.ResponseWriter, r *http.Request) {
	userId := r.Header.Get("x-key")
	params := context.Get(r, "params").(httprouter.Params)
	searchTerm := params.ByName("term")

	detailedResponseUsers, err := c.searchUsers(searchTerm, userId)
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

	err := c.addFeedback(feedback.Feedback, userId)
	if err != nil {
		log.Println("Error adding feedback: ", err)
		WriteError(w, ErrInternalServer)
		return
	}

	Respond(w, r, 204, nil)
}

func (c *appContext) postDeviceTokenHandler(w http.ResponseWriter, r *http.Request) {
	params := context.Get(r, "params").(httprouter.Params)
	userId := r.Header.Get("x-key")
	deviceToken := params.ByName("token")

	err := c.addDeviceToken(deviceToken, userId)
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

	err := c.addPhoneNumber(phoneNumber.PhoneNumber, userId)
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

	err := c.removeDeviceToken(userId)
	if err != nil {
		log.Println("Error removing device token on logout: ", err)
		WriteError(w, ErrInternalServer)
		return
	}

	Respond(w, r, 204, nil)
}
