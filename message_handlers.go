package main

import (
	"log"
	"net/http"

	"github.com/gorilla/context"
)

func (c *appContext) postMessageHandler(w http.ResponseWriter, r *http.Request) {
	userId := r.Header.Get("x-key")
	body := context.Get(r, "body").(*MessageResource)
	message := body.Data

	var newMessageId int
	c.db.Get(&newMessageId, "INSERT INTO messages (from_user_id, message, type) VALUES ($1, $2, $3) RETURNING id", userId, message.Message, message.Type)

	log.Println("messageId: ", newMessageId)
	tx := c.db.MustBegin()
	for _, toUserId := range message.ToUsers {
		log.Println("toUser: ", toUserId)
		tx.MustExec("INSERT INTO message_to_users (user_id, message_id) VALUES ($1, $2)", toUserId, newMessageId)
	}
	if tx.Commit() != nil {
		WriteError(w, ErrInternalServer)
		return
	}

	Respond(w, r, 204, nil)
}

func (c *appContext) getMessagesHandler(w http.ResponseWriter, r *http.Request) {
	userId := r.Header.Get("x-key")

	var newMessageId int
	c.db.Get(&newMessageId, "INSERT INTO messages (from_user_id, message, type) VALUES ($1, $2, $3) RETURNING id", userId, message.Message, message.Type)

	log.Println("messageId: ", newMessageId)
	tx := c.db.MustBegin()
	for _, toUserId := range message.ToUsers {
		log.Println("toUser: ", toUserId)
		tx.MustExec("INSERT INTO message_to_users (user_id, message_id) VALUES ($1, $2)", toUserId, newMessageId)
	}
	if tx.Commit() != nil {
		WriteError(w, ErrInternalServer)
		return
	}

	Respond(w, r, 204, nil)
}
