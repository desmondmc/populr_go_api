package main

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/context"
	"github.com/julienschmidt/httprouter"
)

func (c *appContext) getUsersHandler(w http.ResponseWriter, r *http.Request) {
	var users []User
	c.db.Find(&users)

	w.Header().Set("Content-Type", "application/vnd.api+json")
	json.NewEncoder(w).Encode(users)
}

func (c *appContext) getUserHandler(w http.ResponseWriter, r *http.Request) {
	params := context.Get(r, "params").(httprouter.Params)
	var user User
	c.db.First(&user, params.ByName("id"))

	w.Header().Set("Content-Type", "application/vnd.api+json")
	json.NewEncoder(w).Encode(user)
}

func (c *appContext) getUserFollowersHandler(w http.ResponseWriter, r *http.Request) {
	userId := context.Get(r, "xkey")
	var user User
	c.db.First(&user, userId)

	w.Header().Set("Content-Type", "application/vnd.api+json")
	json.NewEncoder(w).Encode(user.Followers)
}

func (c *appContext) followUserHandler(w http.ResponseWriter, r *http.Request) {
	params := context.Get(r, "params").(httprouter.Params)
	var userToFollow User
	var user User

	c.db.First(&userToFollow, params.ByName("userid"))
	c.db.First(&user, context.Get(r, "xkey"))

	if c.db.Update(&userToFollow).Error != nil {
		WriteError(w, ErrInternalServer)
		return
	}

	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(user.Followers)
}

func (c *appContext) createUserHandler(w http.ResponseWriter, r *http.Request) {
	body := context.Get(r, "body").(*UserResource)
	user := body.Data
	var existingUser User

	c.db.Where("username = ?", user.Username).First(&existingUser)
	if existingUser.ID != 0 {
		WriteError(w, ErrUserExists)
		return
	}

	if c.db.Create(&user).Error != nil {
		WriteError(w, ErrInternalServer)
		return
	}

	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(user)
}
