package main

import (
	"time"

	"github.com/guregu/null/zero"
)

type Public interface {
	Public() interface{}
}

type Resource struct {
	Data interface{} `json:"data"`
}

/************ User Model **************/

type User struct {
	Id       int64  `db:"id" json:"id"`
	Username string `db:"username" json:"username"`
}

type PasswordUser struct {
	User
	Password string `db:"password" json:"password"`
}

type PhoneTokenUser struct {
	User
	PhoneNumber zero.String `db:"phone_number" json:"phone_number"`
	NewToken    string      `db:"new_token" json:"new_token"`
}

type DeviceTokenUser struct {
	User
	Token string `db:"device_token"`
}

type DetailResponseUser struct {
	User
	Friends bool `json:"friends"`
}

type RecieveUserResource struct {
	Data PasswordUser `json:"data"`
}

type PhoneNumberResource struct {
	PhoneNumber string `db:"phone_number" json:"phone_number"`
}

type RecievePhoneNumberResource struct {
	Data PhoneNumberResource `json:"data"`
}

/************ Message Model **************/

type Message struct {
	Id        int64     `db:"id" json:"id"`
	Message   string    `db:"message" json:"message"`
	Type      string    `db:"type" json:"type"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type RecieveMessage struct {
	Message
	ToUsers []int64 `json:"to_users"`
}

type ResponseMessage struct {
	Message
	FromUsername string `db:"username" json:"from_username"`
}

type RecieveMessageResource struct {
	Data RecieveMessage `json:"data"`
}

/************ Feedback Model **************/

type Feedback struct {
	Id       int64  `db:"id"`
	Feedback string `db:"feedback" json:"feedback"`
}

type RecieveFeedbackResource struct {
	Data Feedback `json:"data"`
}
