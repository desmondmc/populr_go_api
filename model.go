package main

type User struct {
	Id       int64  `db:"id" json:"id"`
	Username string `db:"username" json:"username"`
	Password string `db:"password" json:"password"`
}

type UserResource struct {
	Data User `json:"data"`
}

type UsersResource struct {
	Data []User `json:"data"`
}
