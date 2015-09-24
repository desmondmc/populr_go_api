package main

import "time"

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

type RecieveUser struct {
	User
	Password string `db:"password" json:"password"`
}

type ResponseUser struct {
	User
}

type TokenUser struct {
	User
	Token string `db:"device_token"`
}

type DetailResponseUser struct {
	User
	Friends bool `json:"friends"`
}

type RecieveUserResource struct {
	Data RecieveUser `json:"data"`
}

type PhoneNumberResource struct {
	PhoneNumber string `db:"phone_number" json:"phone_number"`
	CountryCode string `db:"country_code" json:"country_code"`
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
