package main

import (
	"database/sql"

	"golang.org/x/crypto/bcrypt"

	"github.com/nu7hatch/gouuid"
)

/************************* Public *************************/

func (c *appContext) getAllUsers(userId string) ([]DetailResponseUser, error) {
	var users []User
	err := c.db.Select(&users, "SELECT id, username FROM users ORDER BY username ASC")
	if err != nil {
		return nil, err
	}

	detailedResponseUsers, err := c.makeDetailResponseUsers(&users, userId)
	if err != nil {
		return nil, err
	}

	return detailedResponseUsers, nil
}

func (c *appContext) getUserWithId(userId string) (*DetailResponseUser, error) {
	var user User
	err := c.db.Get(&user, "SELECT id, username FROM users WHERE id=$1", userId)
	if err == sql.ErrNoRows {
		return nil, err
	}
	detailedResponseUsers, err := c.makeDetailResponseUsers(&[]User{user}, userId)
	if err != nil {
		return nil, err
	}

	return &detailedResponseUsers[0], nil
}

func (c *appContext) getUserWithUsername(username string) (*PasswordUser, error) {
	var savedUser PasswordUser
	err := c.db.Get(&savedUser, "SELECT id, username, password FROM users WHERE username=$1", username)
	return &savedUser, err
}

func (c *appContext) getPhoneTokenUserWithUsername(username string) (*PhoneTokenUser, error) {
	var userToReturn PhoneTokenUser
	err := c.db.Get(&userToReturn, "SELECT id, username, phone_number, new_token FROM users WHERE username=$1", username)
	return &userToReturn, err
}

func (c *appContext) createUser(username, password string) error {
	// Generate Hash From Password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Create UUID for auth
	uuid, err := uuid.NewV4()
	if err != nil {
		return err
	}

	_, err = c.db.Exec(
		"INSERT INTO users (username, password, new_token) VALUES ($1, $2, $3)",
		username,
		string(hashedPassword),
		uuid.String(),
	)
	return err
}

func (c *appContext) addDefaultMessages(toUserId string) {
	messageType := "direct"
	message1 := "Welcome to POPULR! the fastest messaging app in the world! ✋🏽 ✊🏿 👌🏻 ✌🏾 👍🏿 👊🏽 "
	message2 := "Try out emojis on this thing, they come alive, watch: 🌜_______🌛 🌜______🌛 🌜_____🌛 🌜____🌛 🌜___🌛 🌜__🌛 🌜_🌛 🌜🌛 🌝 🌝 🌝 🌖 🌗 🌘 🌚 🌚 🌚 "
	message3 := "Pay attention when you read these messages, because when they’re gone… they’re gone. 💣3 💣2 💣1 💥"

	var message1Id int
	c.db.Get(
		&message1Id,
		"INSERT INTO messages (from_user_id, message, type) VALUES ($1, $2, $3) RETURNING id",
		PopulrUserId,
		message1,
		messageType,
	)
	var message2Id int
	c.db.Get(
		&message2Id,
		"INSERT INTO messages (from_user_id, message, type) VALUES ($1, $2, $3) RETURNING id",
		PopulrUserId,
		message2,
		messageType,
	)
	var message3Id int
	c.db.Get(
		&message3Id,
		"INSERT INTO messages (from_user_id, message, type) VALUES ($1, $2, $3) RETURNING id",
		PopulrUserId,
		message3,
		messageType,
	)

	tx := c.db.MustBegin()
	tx.Exec(
		"INSERT INTO message_to_users (user_id, message_id) VALUES ($1, $2)",
		toUserId,
		message1Id,
	)
	tx.Exec(
		"INSERT INTO message_to_users (user_id, message_id) VALUES ($1, $2)",
		toUserId,
		message2Id,
	)
	tx.Exec(
		"INSERT INTO message_to_users (user_id, message_id) VALUES ($1, $2)",
		toUserId,
		message3Id,
	)
	tx.Commit()
}

/************************* Helpers *************************/

// Given a set of users returns
func (c *appContext) makeDetailResponseUsers(users *[]User, userId string) (responseUsers []DetailResponseUser, err error) {
	passedUsers := *users
	var friends []User

	err = c.db.Select(&friends, findUserFriends, userId)
	if err != nil {
		return nil, err
	}
	userCount := len(passedUsers)
	responseUsers = make([]DetailResponseUser, userCount, userCount)

	// TODO Might be able to make this more efficient.
	for pIndex, passedUser := range passedUsers {
		responseUsers[pIndex].Id = passedUser.Id
		responseUsers[pIndex].Username = passedUser.Username
		for _, friend := range friends {
			if friend.Id == passedUser.Id {
				responseUsers[pIndex].Friends = true
				break
			}
		}
	}

	return responseUsers, nil
}
