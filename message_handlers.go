package main

import (
	"log"
	"net/http"

	"github.com/desmondmcnamee/populr_go_api/Godeps/_workspace/src/github.com/gorilla/context"
	"github.com/desmondmcnamee/populr_go_api/Godeps/_workspace/src/github.com/julienschmidt/httprouter"
)

func (c *appContext) postMessageHandler(w http.ResponseWriter, r *http.Request) {
	userId := r.Header.Get("x-key")
	body := context.Get(r, "body").(*RecieveMessageResource)
	message := body.Data

	var newMessageId int
	err := c.db.Get(&newMessageId, "INSERT INTO messages (from_user_id, message, type) VALUES ($1, $2, $3) RETURNING id", userId, message.Message.Message, message.Message.Type)
	if err != nil {
		log.Println("Error: ", err)
		WriteError(w, ErrInternalServer)
		return
	}

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

const findUserMessages = `
SELECT messages.id, message, type, created_at, username FROM messages 
JOIN message_to_users
ON message_to_users.message_id=messages.id
JOIN users
ON users.id=messages.from_user_id 
WHERE message_to_users.user_id=$1
AND message_to_users.read=FALSE
`

func (c *appContext) getMessagesHandler(w http.ResponseWriter, r *http.Request) {
	userId := r.Header.Get("x-key")

	var messages []ResponseMessage
	err := c.db.Select(&messages, findUserMessages, userId)
	if err != nil {
		log.Println("Error: ", err)
		WriteError(w, ErrInternalServer)
		return
	}

	Respond(w, r, 201, messages)
}

const markAsRead = `
UPDATE message_to_users
SET read = TRUE 
WHERE message_id = $1
AND user_id = $2
`

func (c *appContext) readMessageHandler(w http.ResponseWriter, r *http.Request) {
	userId := r.Header.Get("x-key")
	params := context.Get(r, "params").(httprouter.Params)
	messageId := params.ByName("id")

	_, err := c.db.Exec(markAsRead, messageId, userId)
	if err != nil {
		log.Println("Error: ", err)
		WriteError(w, ErrInternalServer)
		return
	}

	Respond(w, r, 204, nil)
}
