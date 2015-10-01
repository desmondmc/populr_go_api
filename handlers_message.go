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

	newMessageId, err := c.addMessage(userId, message.Message.Message, message.Type)
	if err != nil {
		log.Println("Error adding message: ", err)
		WriteError(w, ErrInternalServer)
		return
	}

	toUsersIds, err := c.getToUsersForMessageType(message.Type, userId, message)
	if err != nil {
		log.Println("Error: ", err)
		WriteError(w, ErrInternalServer)
		return
	}

	err = c.addMessageToUsers(toUsersIds, newMessageId)
	if err != nil {
		log.Println("Error adding toUsers", err)
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

func (c *appContext) getMessagesHandler(w http.ResponseWriter, r *http.Request) {
	userId := r.Header.Get("x-key")

	messages, err := c.getUserMessages(userId)
	if err != nil {
		log.Println("Error getting user messages: ", err)
		WriteError(w, ErrInternalServer)
		return
	}

	Respond(w, r, 201, messages)
}

func (c *appContext) readMessageHandler(w http.ResponseWriter, r *http.Request) {
	userId := r.Header.Get("x-key")
	params := context.Get(r, "params").(httprouter.Params)
	messageId := params.ByName("id")

	err := c.markMessageAsRead(userId, messageId)
	if err != nil {
		log.Println("Error: ", err)
		WriteError(w, ErrInternalServer)
		return
	}

	Respond(w, r, 204, nil)
}
