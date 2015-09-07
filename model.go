package main

import "time"

type User struct {
	ID        uint       `json:"id" gorm:"primary_key"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
	Username  string     `json:"username" sql:"type:varchar(100);unique_index"`
	Password  string     `json:"password"`
}

type UserResource struct {
	Data User `json:"data"`
}
