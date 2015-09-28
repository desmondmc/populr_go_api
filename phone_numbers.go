package main

type PhoneUser struct {
	User
	PhoneNumber string `db:"phone_number" json:"phone_number"`
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
	query := "SELECT username, id FROM users WHERE phone_number = 'DOG'"
	for _, contact := range contacts {
		for _, number := range contact.Phones {
			if number == "" {
				continue
			}
			query = query + " OR phone_number = '" + number + "'"
		}
	}

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
