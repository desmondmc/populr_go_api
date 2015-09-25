package main

import (
	"log"
	"strconv"

	"github.com/ttacon/libphonenumber"
)

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

func getRawNumber(number, cc string) string {
	num, _ := libphonenumber.Parse(number, cc)

	countryCode := strconv.Itoa(int(num.GetCountryCode()))
	nationalNumer := strconv.Itoa(int(num.GetNationalNumber()))
	fullNumber := countryCode + nationalNumer
	log.Println("Full Number: ", fullNumber)
	return fullNumber
}

// Find usernames with numbers that match one of the numbers in the contact array.
func (c *appContext) processContacts(contacts []Contact, cc string) []PhoneUser {
	// Build giant query. 'DOG' is there to make the query building easier.
	query := "SELECT username, id, phone_number FROM users WHERE phone_number = 'DOG'"
	for _, contact := range contacts {
		for _, number := range contact.Phones {
			number = getRawNumber(number, cc)
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
