package main

import (
	"fmt"
	"log"

	"github.com/desmondmcnamee/populr_go_api/Godeps/_workspace/src/github.com/anachronistic/apns"
)

func (c *appContext) SendNewPublicMessagePush(userIds []int64) {
	c.sendNewMessagePush(userIds, "New message")
}

func (c *appContext) SendNewDirectMessagePush(userIds []int64) {
	c.sendNewMessagePush(userIds, "New direct message")
}

func (c *appContext) SendNewFriendPush(userId string) {
	c.sendPushWithIdAndMessage(userId, "New friend!")
}

func (c *appContext) sendNewMessagePush(userIds []int64, message string) {
	for _, userId := range userIds {
		idString := fmt.Sprintf("%d", userId)
		c.sendPushWithIdAndMessage(idString, message)
	}
}

func (c *appContext) sendPushWithIdAndMessage(id, message string) {
	log.Println("Push data - id: ", id)
	log.Println("id: ", id)
	log.Println("message: ", message)

	var tokenUser TokenUser
	err := c.db.Get(&tokenUser, "SELECT id, username FROM users WHERE id=$1", id)
	if err != nil {
		log.Println("Error finding user for push: ", err)
	}

	log.Println("Sending push to user: ", tokenUser)

	if tokenUser.Token != "" {
		sendPush(
			tokenUser.Token,
			message,
			"new_message",
		)
	}
}

func sendPush(token, message, mtype string) {
	go func() {
		payload := apns.NewPayload()
		payload.Alert = message
		payload.Badge = 1

		pn := apns.NewPushNotification()
		pn.DeviceToken = token
		pn.AddPayload(payload)

		pn.Set("type", mtype)

		client := apns.NewClient(PRO_SERVER, CERT_PEM, KEY_PEM)
		resp := client.Send(pn)

		alert, _ := pn.PayloadString()
		log.Println("Sending Push: ")
		log.Println("  Alert:", alert)
		log.Println("Success:", resp.Success)
		log.Println("  Error:", resp.Error)
	}()
}

const CERT_PEM = "./pop-prod-cert.pem"

const KEY_PEM = "./pop-prod-key-noenc.pem"

const DEV_SERVER = "gateway.sandbox.push.apple.com:2195"
const PRO_SERVER = "gateway.push.apple.com:2195"
