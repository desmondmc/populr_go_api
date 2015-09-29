package main

import "github.com/desmondmcnamee/populr_go_api/Godeps/_workspace/src/github.com/guregu/null/zero"

type PhoneUser struct {
	User
	PhoneNumber zero.String `db:"phone_number" json:"phone_number"`
	NewToken    string      `db:"new_token" json:"new_token"`
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
func (c *appContext) processContacts(contacts []Contact, userId string) ([]DetailResponseUser, error) {
	// Build giant query. 'DOG' is there to make the query building easier.
	query := "SELECT username, id FROM users WHERE (phone_number ='DOG'"
	for _, contact := range contacts {
		for _, number := range contact.Phones {
			if number == "" {
				continue
			}
			query = query + " OR phone_number = '" + number + "'"
		}
	}

	// Prevents returning to the user themself as a suggestion.
	query = query + ") AND id != " + userId

	var users []ResponseUser
	err := c.db.Select(&users, query)
	if err != nil {
		return nil, err
	}

	detailResponseUsers, err := c.MakeDetailResponseUsers(&users, userId)
	if err != nil {
		return nil, err
	}

	return detailResponseUsers, nil
}
