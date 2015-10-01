package main

import (
	"log"
	"net/http"

	"github.com/desmondmcnamee/populr_go_api/Godeps/_workspace/src/github.com/gorilla/context"
	"github.com/desmondmcnamee/populr_go_api/Godeps/_workspace/src/github.com/julienschmidt/httprouter"
)

// Followers

func (c *appContext) getUserFriendsHandler(w http.ResponseWriter, r *http.Request) {
	userId := r.Header.Get("x-key")

	detailedResponseUsers, err := c.getUserFriends(userId)
	if err != nil {
		log.Println("Error finding user friends: ", err)
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

func (c *appContext) unfriendUserHandler(w http.ResponseWriter, r *http.Request) {
	params := context.Get(r, "params").(httprouter.Params)

	userIdToUnfriend := params.ByName("id")
	userId := r.Header.Get("x-key")

	if c.alreadyFriendsCheck(userId, userIdToUnfriend) == false {
		WriteError(w, ErrNotFriends)
		return
	}

	err := c.deleteFriend(userId, userIdToUnfriend)
	if err != nil {
		log.Println("Error removing friend ", err)
		WriteError(w, ErrInternalServer)
		return
	}

	Respond(w, r, 204, nil)
}
