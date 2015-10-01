package main

import (
	"errors"
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

	toUsersIds, err := c.getToUsersForMessageType(message.Type, userId, message)
	if err != nil {
		log.Println("Error: ", err)
		WriteError(w, ErrInternalServer)
		return
	}

	tx := c.db.MustBegin()
	for _, toUserId := range toUsersIds {
		log.Println("toUser: ", toUserId)
		tx.MustExec("INSERT INTO message_to_users (user_id, message_id) VALUES ($1, $2)", toUserId, newMessageId)
	}
	if tx.Commit() != nil {
		WriteError(w, ErrInternalServer)
		return
	}

	Respond(w, r, 204, nil)

	// Send pushes
	if message.Type == "direct" {
		c.SendNewDirectMessagePush(toUsersIds)
	}
	if message.Type == "public" {
		c.SendNewPublicMessagePush(toUsersIds)
	}
}

const findUserMessages = `
SELECT messages.id, message, type, created_at, username FROM messages 
JOIN message_to_users
ON message_to_users.message_id=messages.id
JOIN users
ON users.id=messages.from_user_id 
WHERE message_to_users.user_id=$1
AND message_to_users.read=FALSE
ORDER BY created_at DESC
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

/************* HELPERS ****************/

func (c *appContext) getToUsersForMessageType(messageType, userId string, message RecieveMessage) ([]int64, error) {
	var users []User

	if messageType == "public" {
		var err error

		// Check if this is the populr user.
		if userId == PopulrUserId {
			// Get all users.
			err = c.db.Select(&users, "SELECT users.id, users.username FROM users")
		} else {
			// Get all user's friends IDs those are the ToUsers
			err = c.db.Select(&users, findUserFriends, userId)
		}

		if err != nil {
			return nil, err
		}

		var toUserIds []int64
		userCount := len(users)
		toUserIds = make([]int64, userCount, userCount)
		for index, toUser := range users {
			toUserIds[index] = toUser.Id
		}
		return toUserIds, nil
	}

	if messageType == "direct" {
		return message.ToUsers, nil
	}

	return nil, errors.New("Invalid message type")
}
