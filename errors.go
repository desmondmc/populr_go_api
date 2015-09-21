package main

import (
	"encoding/json"
	"net/http"
)

type Errors struct {
	Errors []*Error `json:"errors"`
}

type Error struct {
	Id      string `json:"id"`
	Status  int    `json:"status"`
	Title   string `json:"title"`
	Detail  string `json:"detail"`
	Message string `json:"message"` // User friendly message
}

func WriteError(w http.ResponseWriter, err *Error) {
	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(err.Status)
	json.NewEncoder(w).Encode(Errors{[]*Error{err}})
}

var (
	ErrBadRequest           = &Error{"bad_request", 400, "Bad request", "Request body is not well-formed. It must be JSON.", ""}
	ErrNotAcceptable        = &Error{"not_acceptable", 406, "Not Acceptable", "Accept header must be set to 'application/vnd.api+json'.", ""}
	ErrUnsupportedMediaType = &Error{"unsupported_media_type", 415, "Unsupported Media Type", "Content-Type header must be set to: 'application/vnd.api+json'.", ""}
	ErrInternalServer       = &Error{"internal_server_error", 500, "Internal Server Error", "Something went wrong.", "Something went wrong."}
	ErrUserExists           = &Error{"user_already_exists", 409, "User Already Exists", "New users must have unique username.", "Username taken. Sorry"}
	ErrNotFriends           = &Error{"not_friends", 409, "Not Friends with User", "Tried to unfriend a user who you are not friends with.", ""}
	ErrCannotFriendSelf     = &Error{"cannot_friend_self", 409, "Can't friend yourself", "You tried to friend youself. Don't do that.", ""}
	ErrAlreadyFriends       = &Error{"already_friends", 409, "Can't befriend someone twice", "You tried to friend someone twice. Don't do that.", ""}
	ErrFriending            = &Error{"friending_error", 409, "Friending Error", "Either you or the the person you are trying to friend do not exist. Strange.", ""}
	ErrNoXKey               = &Error{"no_x_key", 409, "No x-key value in header", "HTTP x-key needs to be set for this request.", ""}
	ErrInvalidLogin         = &Error{"invalid_login", 409, "Invalid Login", "The username or password is incorrect.", "The username or password is incorrect."}
	ErrNoUserForId          = &Error{"no_user_for_id", 409, "Could not find user", "No user found for that Id.", ""}
)
