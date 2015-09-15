package main

import "time"

type User struct {
	ID        uint       `json:"id" gorm:"primary_key"`
	CreatedAt time.Time  `json:"_"`
	UpdatedAt time.Time  `json:"_"`
	DeletedAt *time.Time `json:"_"`
	Username  string     `json:"username" sql:"type:varchar(100);unique_index"`
	Password  string     `json:"_"`
	Followers []User     `gorm:"foreignkey:user_id;associationforeignkey:follower_id;many2many:user_followers;" json:"followers"`
}

type UserResource struct {
	Data User `json:"data"`
}
