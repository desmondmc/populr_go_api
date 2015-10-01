package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"

	"github.com/nu7hatch/gouuid"
)

// Some long queries are stored in constants to make them more readable.

const findUserFriends = `
SELECT users.id, users.username FROM users 
JOIN friends 
ON friends.friend_id=users.id 
WHERE friends.user_id=$1
`

const friendsWithUserCheck = `
SELECT users.id, users.username FROM users 
JOIN friends 
ON friends.friend_id=users.id 
WHERE friends.friend_id=$1
AND friends.user_id=$2
`

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

const markAsRead = `
UPDATE message_to_users
SET read = TRUE 
WHERE message_id = $1
AND user_id = $2
`

/************************* Public *************************/

func (c *appContext) validateAuthTokenForUser(userId, token string) error {
	var user PhoneTokenUser
	err := c.db.Get(&user, "SELECT id, username, phone_number, new_token FROM users WHERE id=$1", userId)
	if err != nil {
		return err
	}

	if token != user.NewToken {
		err = errors.New("AuthToken mismatch")
	}
	return err
}

func (c *appContext) markMessageAsRead(userId, messageId string) error {
	_, err := c.db.Exec(markAsRead, messageId, userId)
	return err
}

func (c *appContext) getUserMessages(userId string) ([]ResponseMessage, error) {
	var messages []ResponseMessage
	err := c.db.Select(&messages, findUserMessages, userId)
	return messages, err
}

func (c *appContext) addMessageToUsers(toUserIds []int64, messageId int) error {
	tx := c.db.MustBegin()
	for _, toUserId := range toUserIds {
		log.Println("toUser: ", toUserId)
		tx.MustExec("INSERT INTO message_to_users (user_id, message_id) VALUES ($1, $2)", toUserId, messageId)
	}
	return tx.Commit()
}

// Returns the message id of the new message.
func (c *appContext) addMessage(userId, message, messageType string) (int, error) {
	var newMessageId int
	err := c.db.Get(
		&newMessageId,
		"INSERT INTO messages (from_user_id, message, type) VALUES ($1, $2, $3) RETURNING id",
		userId,
		message,
		messageType,
	)
	return newMessageId, err
}

func (c *appContext) alreadyFriendsCheck(userId, friendId string) bool {
	var userToCheck User
	c.db.Get(&userToCheck, friendsWithUserCheck, userId, friendId)

	log.Println("userId: ", userId, "friendId", friendId)

	if userToCheck.Id != 0 {
		return true
	}

	return false
}

func (c *appContext) deleteFriend(sourceId, targetId string) error {
	tx := c.db.MustBegin()
	c.db.MustExec("DELETE FROM friends WHERE friends.user_id=$1 AND friends.friend_id=$2", targetId, sourceId)
	c.db.MustExec("DELETE FROM friends WHERE friends.user_id=$1 AND friends.friend_id=$2", sourceId, targetId)
	return tx.Commit()
}

func (c *appContext) addFriend(sourceId, targetId string) (err error) {
	var userToFriend User
	var user User

	err = c.db.Get(&userToFriend, "SELECT id, username FROM users WHERE id=$1", targetId)
	err = c.db.Get(&user, "SELECT id, username FROM users WHERE id=$1", sourceId)

	if err != nil || userToFriend.Id == 0 || user.Id == 0 {
		log.Println("Error finding user: ", err)
		return err
	}

	tx := c.db.MustBegin()
	tx.MustExec("INSERT INTO friends (user_id, friend_id) VALUES ($1, $2)", user.Id, userToFriend.Id)
	tx.MustExec("INSERT INTO friends (user_id, friend_id) VALUES ($1, $2)", userToFriend.Id, user.Id)
	if tx.Commit() != nil {
		return errors.New("error inserting friend.")
	}

	return nil
}

func (c *appContext) getUserFriends(userId string) ([]DetailResponseUser, error) {
	var users []User
	err := c.db.Select(&users, findUserFriends, userId)
	detailedResponseUsers, err := c.makeDetailResponseUsers(&users, userId)
	if err != nil {
		return nil, err
	}

	return detailedResponseUsers, nil
}

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

func (c *appContext) searchUsers(term, excludeId string) ([]DetailResponseUser, error) {
	term = fmt.Sprint(term, "%")

	var users []User
	c.db.Select(
		&users,
		"SELECT id, username FROM users WHERE users.username LIKE $1 AND users.id != $2 ORDER BY username ASC",
		term,
		excludeId,
	)

	detailedResponseUsers, err := c.makeDetailResponseUsers(&users, excludeId)

	return detailedResponseUsers, err
}

func (c *appContext) addFeedback(feedback, userId string) error {
	_, err := c.db.Exec("INSERT INTO feedbacks (user_id, feedback) VALUES ($1, $2)",
		userId,
		feedback,
	)
	return err
}

func (c *appContext) addDeviceToken(deviceToken, userId string) error {
	_, err := c.db.Exec("UPDATE users SET device_token = $1 WHERE id = $2",
		deviceToken,
		userId,
	)
	return err
}

func (c *appContext) addPhoneNumber(phoneNumber, userId string) error {
	_, err := c.db.Exec(
		"UPDATE users SET phone_number = $1 WHERE id = $2",
		phoneNumber,
		userId,
	)
	return err
}

func (c *appContext) removeDeviceToken(userId string) error {
	_, err := c.db.Exec("UPDATE users SET device_token = $1 WHERE id = $2",
		"",
		userId,
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
