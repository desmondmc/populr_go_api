package main

import "log"

type PhoneUser struct {
	User
	PhoneNumber string `json:"phone_number"`
}

type Contact struct {
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	Phones    []string `json:"phones"`
}

type RecieveContacts struct {
	Data []Contact `json:"data"`
}

// Find usernames with numbers that match one of the numbers in the contact array.
func (c *appContext) processContacts(contacts []Contact) []PhoneUser {
	// Build giant query. 'DOG' is there to make the query building easier.
	query := "SELECT username, id, phone_number FROM users WHERE phone_number = 'DOG'"
	for _, contact := range contacts {
		for _, number := range contact.Phones {
			if number == "" {
				continue
			}
			query = query + "OR phone_number = " + number
		}
	}

	log.Println("&&&&&&&&&&&&&&&&&\n", query)

	var users []PhoneUser
	c.db.Select(users, query)

	return users
}