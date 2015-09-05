package main

import "gopkg.in/mgo.v2/bson"

type User struct {
	Id       bson.ObjectId `json:"id,omitempty" bson:"_id,omitempty"`
	Username string        `json:"username"`
	Password string        `json:"password"`
}

type UsersCollection struct {
	Data []User `json:"data"`
}

type UserResource struct {
	Data User `json:"data"`
}
