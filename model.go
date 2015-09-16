package main

type Public interface {
	Public() interface{}
}

type User struct {
	Id       int64  `db:"id" json:"id"`
	Username string `db:"username" json:"username"`
	Password string `db:"password" json:"password"`
}

type Resource struct {
	Data interface{} `json:"data"`
}

// This structure is used for decoding recieved user data json
type UserResource struct {
	Data User `json:"data"`
}

// Used as a wrapper to implement Public method when returning multiple users.
type UsersResource struct {
	Users []User
}

func (u User) Public() interface{} {
	return map[string]interface{}{
		"id":       u.Id,
		"username": u.Username,
	}
}

func (users UsersResource) Public() interface{} {
	numUsers := len(users.Users)
	publicUsers := make([]interface{}, numUsers, numUsers)

	for i, user := range users.Users {
		publicUsers[i] = user.Public()
	}

	return publicUsers
}
