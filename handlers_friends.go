package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/gorilla/context"
	"github.com/julienschmidt/httprouter"
)

// Followers

const findUserFriends = `
SELECT users.id, users.username FROM users 
JOIN friends 
ON friends.friend_id=users.id 
WHERE friends.user_id=$1
`

// Given a set of users returns
func (c *appContext) MakeDetailResponseUsers(users *[]ResponseUser, userId string) (responseUsers []DetailResponseUser, err error) {
	passedUsers := *users
	var friends []User

	err = c.db.Select(&friends, findUserFriends, userId)
	if err != nil {
		return nil, err
	}
	userCount := len(passedUsers)
	responseUsers = make([]DetailResponseUser, userCount, userCount)

	// TODO Might be able to make this more efficient.
	for pIndex, passedUser := range passedUsers {
		responseUsers[pIndex].Id = passedUser.Id
		responseUsers[pIndex].Username = passedUser.Username
		for _, friend := range friends {
			if friend.Id == passedUser.Id {
				responseUsers[pIndex].Friends = true
				break
			}
		}
	}

	return responseUsers, nil
}

func (c *appContext) getUserFriendsHandler(w http.ResponseWriter, r *http.Request) {
	userId := r.Header.Get("x-key")
	var users []ResponseUser

	c.db.Select(&users, findUserFriends, userId)

	detailedResponseUsers, err := c.MakeDetailResponseUsers(&users, userId)
	if err != nil {
		log.Println("Error searching on users: ", err)
		WriteError(w, ErrInternalServer)
		return
	}

	Respond(w, r, 201, detailedResponseUsers)
}

func (c *appContext) friendUserHandler(w http.ResponseWriter, r *http.Request) {
	params := context.Get(r, "params").(httprouter.Params)

	userToFriendId := params.ByName("id")
	userId := r.Header.Get("x-key")

	if userToFriendId == userId {
		WriteError(w, ErrCannotFriendSelf)
		return
	}

	if c.alreadyFriendsCheck(userId, userToFriendId) == true {
		WriteError(w, ErrAlreadyFriends)
		return
	}

	err := c.addFriend(userId, userToFriendId)
	if err != nil {
		WriteError(w, ErrInternalServer)
		log.Println("Error friending user: ", err)
		return
	}

	Respond(w, r, 204, nil)

	c.SendNewFriendPush(userToFriendId, userId)
}

func (c *appContext) addFriend(sourceId, targetId string) (err error) {
	var userToFriend User
	var user User

	err = c.db.Get(&userToFriend, "SELECT id, username FROM users WHERE id=$1", targetId)
	err = c.db.Get(&user, "SELECT id, username FROM users WHERE id=$1", sourceId)

	if err != nil || userToFriend.Id == 0 || user.Id == 0 {
		log.Println("Error finding user: ", err)
		return err
	}

	tx := c.db.MustBegin()
	tx.MustExec("INSERT INTO friends (user_id, friend_id) VALUES ($1, $2)", user.Id, userToFriend.Id)
	tx.MustExec("INSERT INTO friends (user_id, friend_id) VALUES ($1, $2)", userToFriend.Id, user.Id)
	if tx.Commit() != nil {
		return errors.New("error inserting friend.")
	}

	return nil
}

func (c *appContext) unfriendUserHandler(w http.ResponseWriter, r *http.Request) {
	params := context.Get(r, "params").(httprouter.Params)

	userIdToUnfriend := params.ByName("id")
	userId := r.Header.Get("x-key")

	if c.alreadyFriendsCheck(userId, userIdToUnfriend) == false {
		WriteError(w, ErrNotFriends)
		return
	}

	tx := c.db.MustBegin()
	c.db.MustExec("DELETE FROM friends WHERE friends.user_id=$1 AND friends.friend_id=$2", userIdToUnfriend, userId)
	c.db.MustExec("DELETE FROM friends WHERE friends.user_id=$1 AND friends.friend_id=$2", userId, userIdToUnfriend)
	if tx.Commit() != nil {
		WriteError(w, ErrInternalServer)
		return
	}

	Respond(w, r, 204, nil)
}

const friendsWithUserCheck = `
SELECT users.id, users.username FROM users 
JOIN friends 
ON friends.friend_id=users.id 
WHERE friends.friend_id=$1
AND friends.user_id=$2
`

func (c *appContext) alreadyFriendsCheck(userId, friendId string) bool {
	var userToCheck User
	c.db.Get(&userToCheck, friendsWithUserCheck, userId, friendId)

	log.Println("userId: ", userId, "friendId", friendId)

	if userToCheck.Id != 0 {
		return true
	}

	return false
}
