package main

import (
	"encoding/json"
	"net/http"
)

type Errors struct {
	Errors []*Error `json:"errors"`
}

type Error struct {
	Id     string `json:"id"`
	Status int    `json:"status"`
	Title  string `json:"title"`
	Detail string `json:"detail"`
}

func WriteError(w http.ResponseWriter, err *Error) {
	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(err.Status)
	json.NewEncoder(w).Encode(Errors{[]*Error{err}})
}

var (
	ErrBadRequest           = &Error{"bad_request", 400, "Bad request", "Request body is not well-formed. It must be JSON."}
	ErrNotAcceptable        = &Error{"not_acceptable", 406, "Not Acceptable", "Accept header must be set to 'application/vnd.api+json'."}
	ErrUnsupportedMediaType = &Error{"unsupported_media_type", 415, "Unsupported Media Type", "Content-Type header must be set to: 'application/vnd.api+json'."}
	ErrInternalServer       = &Error{"internal_server_error", 500, "Internal Server Error", "Something went wrong."}
	ErrUserExists           = &Error{"user_already_exists", 409, "User Already Exists", "New users must have unique username."}
	ErrNotFollowingUser     = &Error{"not_following_user", 409, "Not Following User", "Tried to unfollow a user who you are not following."}
	ErrCannotFollowSelf     = &Error{"cannot_follow_self", 409, "Can't follow yourself", "You tried to follow youself. Don't do that."}
	ErrAlreadyFollowing     = &Error{"already_following_user", 409, "Can't follow someone twice", "You tried to follow someone twice. Don't do that."}
	ErrFollowing            = &Error{"following_error", 409, "Following Error", "Either you or the the person you are trying to follow do not exist. Strange."}
	ErrNoXKey               = &Error{"no_x_key", 409, "No x-key value in header", "HTTP x-key needs to be set for this request."}
	ErrInvalidLogin         = &Error{"invalid_login", 409, "Invalid Login", "The username or password is incorrect."}
)
