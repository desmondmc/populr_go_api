package main

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/context"
	"github.com/julienschmidt/httprouter"
	"gopkg.in/mgo.v2/bson"
)

func (c *appContext) getUsersHandler(w http.ResponseWriter, r *http.Request) {
	repo := PopulrRepo{c.db.C("users")}
	users, err := repo.AllUsers()
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/vnd.api+json")
	json.NewEncoder(w).Encode(users)
}

func (c *appContext) getUserHandler(w http.ResponseWriter, r *http.Request) {
	params := context.Get(r, "params").(httprouter.Params)
	repo := PopulrRepo{c.db.C("users")}
	user, err := repo.FindUser(params.ByName("id"))
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/vnd.api+json")
	json.NewEncoder(w).Encode(user)
}

func (c *appContext) createUserHandler(w http.ResponseWriter, r *http.Request) {
	body := context.Get(r, "body").(*UserResource)
	repo := PopulrRepo{c.db.C("users")}
	err := repo.CreateUser(&body.Data)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(body)
}

func (c *appContext) updateUserHandler(w http.ResponseWriter, r *http.Request) {
	params := context.Get(r, "params").(httprouter.Params)
	body := context.Get(r, "body").(*UserResource)
	body.Data.Id = bson.ObjectIdHex(params.ByName("id"))
	repo := PopulrRepo{c.db.C("teas")}
	err := repo.UpdateUser(&body.Data)
	if err != nil {
		panic(err)
	}

	w.WriteHeader(204)
	w.Write([]byte("\n"))
}

func (c *appContext) deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	params := context.Get(r, "params").(httprouter.Params)
	repo := PopulrRepo{c.db.C("users")}
	err := repo.Delete(params.ByName("id"))
	if err != nil {
		panic(err)
	}

	w.WriteHeader(204)
	w.Write([]byte("\n"))
}
