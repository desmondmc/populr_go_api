package main

import (
	"log"
	"net/http"

	"github.com/desmondmcnamee/populr_go_api/Godeps/_workspace/src/github.com/gorilla/context"
	"github.com/desmondmcnamee/populr_go_api/Godeps/_workspace/src/github.com/julienschmidt/httprouter"
)

// Followers

const findUserfollowers = `
SELECT users.id, users.username FROM users 
JOIN user_followers 
ON user_followers.follower_id=users.id 
WHERE user_followers.user_id=$1
`

func (c *appContext) getUserFollowersHandler(w http.ResponseWriter, r *http.Request) {
	userId := r.Header.Get("x-key")
	var users []ResponseUser

	log.Println("userId: ", userId)

	c.db.Select(&users, findUserfollowers, userId)

	Respond(w, r, 201, users)
}

const findUsersFollowing = `
SELECT users.id, users.username FROM users 
JOIN user_followers 
ON user_followers.user_id=users.id 
WHERE user_followers.follower_id=$1
`

func (c *appContext) getUsersFollowingHandler(w http.ResponseWriter, r *http.Request) {
	userId := r.Header.Get("x-key")
	var users []ResponseUser

	log.Println("userId: ", userId)

	c.db.Select(&users, findUsersFollowing, userId)

	Respond(w, r, 201, users)
}

func (c *appContext) followUserHandler(w http.ResponseWriter, r *http.Request) {
	params := context.Get(r, "params").(httprouter.Params)

	userToFollowId := params.ByName("id")
	userId := r.Header.Get("x-key")

	if userToFollowId == userId {
		WriteError(w, ErrCannotFollowSelf)
		return
	}

	if c.alreadyFollowingUser(userId, userToFollowId) == true {
		WriteError(w, ErrAlreadyFollowing)
		return
	}

	var userToFollow User
	var user User

	err := c.db.Get(&userToFollow, "SELECT id, username FROM users WHERE id=$1", userToFollowId)
	err = c.db.Get(&user, "SELECT id, username FROM users WHERE id=$1", userId)

	if err != nil || userToFollow.Id == 0 || user.Id == 0 {
		log.Println("Error finding user: ", err)
		WriteError(w, ErrFollowing)
		return
	}

	tx := c.db.MustBegin()
	tx.MustExec("INSERT INTO user_followers (user_id, follower_id) VALUES ($1, $2)", userToFollow.Id, user.Id)
	if tx.Commit() != nil {
		WriteError(w, ErrInternalServer)
		return
	}

	Respond(w, r, 204, nil)
}

func (c *appContext) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	params := context.Get(r, "params").(httprouter.Params)

	userIdToUnfollow := params.ByName("id")
	userId := r.Header.Get("x-key")

	if c.alreadyFollowingUser(userId, userIdToUnfollow) == false {
		WriteError(w, ErrNotFollowingUser)
		return
	}

	tx := c.db.MustBegin()
	c.db.MustExec("DELETE FROM user_followers WHERE user_followers.user_id=$1 AND user_followers.follower_id=$2", userIdToUnfollow, userId)
	if tx.Commit() != nil {
		WriteError(w, ErrInternalServer)
		return
	}

	Respond(w, r, 204, nil)
}

const followingUserCheck = `
SELECT users.id, users.username FROM users 
JOIN user_followers 
ON user_followers.follower_id=users.id 
WHERE user_followers.follower_id=$1
AND user_followers.user_id=$2
`

func (c *appContext) alreadyFollowingUser(userId, followingId string) bool {
	var userToCheck User
	c.db.Get(&userToCheck, followingUserCheck, userId, followingId)

	log.Println("userId: ", userId, "followId", followingId)

	if userToCheck.Id != 0 {
		return true
	}

	return false
}
