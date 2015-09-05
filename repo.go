package main

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Repo

type PopulrRepo struct {
	coll *mgo.Collection
}

func (r *PopulrRepo) AllUsers() (UsersCollection, error) {
	result := UsersCollection{[]User{}}
	err := r.coll.Find(nil).All(&result.Data)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (r *PopulrRepo) FindUser(id string) (UserResource, error) {
	result := UserResource{}
	err := r.coll.FindId(bson.ObjectIdHex(id)).One(&result.Data)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (r *PopulrRepo) CreateUser(user *User) error {
	id := bson.NewObjectId()
	_, err := r.coll.UpsertId(id, user)
	if err != nil {
		return err
	}

	user.Id = id

	return nil
}

func (r *PopulrRepo) UpdateUser(user *User) error {
	err := r.coll.UpdateId(user.Id, user)
	if err != nil {
		return err
	}

	return nil
}

func (r *PopulrRepo) Delete(id string) error {
	err := r.coll.RemoveId(bson.ObjectIdHex(id))
	if err != nil {
		return err
	}

	return nil
}
