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

func (c *appContext) createUserHandler(w http.ResponseWriter, r *http.Request) {
	body := context.Get(r, "body").(*UserResource)
	user := body.Data

	if !c.db.NewRecord(user) {
		WriteError(w, ErrUserExists)
	}

	c.db.Create(&user)

	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(user)
}
