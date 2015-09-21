package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/desmondmcnamee/populr_go_api/Godeps/_workspace/src/github.com/gorilla/context"
	"github.com/desmondmcnamee/populr_go_api/Godeps/_workspace/src/github.com/jmoiron/sqlx"
	"github.com/desmondmcnamee/populr_go_api/Godeps/_workspace/src/github.com/justinas/alice"
	_ "github.com/desmondmcnamee/populr_go_api/Godeps/_workspace/src/github.com/lib/pq"
)

type appContext struct {
	db *sqlx.DB
}

func main() {
	// Get and set env variables
	portString := ":" + os.Getenv("PORT")
	if os.Getenv("PORT") == "" {
		portString = ":8080"
	}
	database := os.Getenv("DATABASE_URL")
	if os.Getenv("DATABASE_URL") == "" {
		database = "user=desmondmcnamee dbname=populr sslmode=disable"
	}

	// Setup database
	db, err := sqlx.Connect("postgres", database)
	if err != nil {
		fmt.Printf("sql.Open error: %v\n", err)
		return
	}
	defer db.Close()

	log.Println("Setting up database..")
	dbSetup(db)

	// Setup middleware
	appC := appContext{db}
	commonHandlers := alice.New(context.ClearHandler, loggingHandler, recoverHandler, acceptHandler)
	loggedInCommonHandlers := commonHandlers.Append(contentTypeHandler)

	// Setup Routes
	log.Println("Setting up routes...")
	router := NewRouter()
	router.Get("/user/:id", commonHandlers.ThenFunc(appC.getUserHandler))
	router.Get("/friends", loggedInCommonHandlers.ThenFunc(appC.getUserFriendsHandler))
	router.Get("/users", commonHandlers.ThenFunc(appC.getUsersHandler))
	router.Get("/searchusers/:term", loggedInCommonHandlers.ThenFunc(appC.searchUsersHandler))
	router.Get("/messages", loggedInCommonHandlers.ThenFunc(appC.getMessagesHandler))
	router.Post("/signup", commonHandlers.Append(contentTypeHandler, bodyHandler(RecieveUserResource{})).ThenFunc(appC.createUserHandler))
	router.Post("/login", commonHandlers.Append(contentTypeHandler, bodyHandler(RecieveUserResource{})).ThenFunc(appC.loginUserHandler))
	router.Post("/friend/:id", loggedInCommonHandlers.ThenFunc(appC.friendUserHandler))
	router.Post("/readmessage/:id", loggedInCommonHandlers.ThenFunc(appC.readMessageHandler))
	router.Post("/message", commonHandlers.Append(contentTypeHandler, bodyHandler(RecieveMessageResource{})).ThenFunc(appC.postMessageHandler))
	router.Delete("/unfriend/:id", loggedInCommonHandlers.ThenFunc(appC.unfriendUserHandler))

	log.Println("Listening...")
	http.ListenAndServe(portString, router)
}

var schema = `
CREATE TABLE users (
	id SERIAL NOT NULL PRIMARY KEY,
    username text,
    password text
);

CREATE TABLE friends (
      user_id    int REFERENCES users (id) ON UPDATE CASCADE ON DELETE CASCADE, 
      friend_id int REFERENCES users (id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE messages (
	id SERIAL NOT NULL PRIMARY KEY,
    from_user_id int REFERENCES users (id) ON UPDATE CASCADE ON DELETE CASCADE,
    message text,
    type text,
    created_at timestamp default now()
);

CREATE TABLE message_to_users (
      user_id    int REFERENCES users (id) ON UPDATE CASCADE ON DELETE CASCADE, 
      message_id int REFERENCES messages (id) ON UPDATE CASCADE ON DELETE CASCADE,
      read bool default false
);
`

var dropAllTables = `
	DROP TABLE message_to_users, user_followers, messages, users;
`

func dbSetup(db *sqlx.DB) {
	db.MustExec(dropAllTables)
	db.Exec(schema)
}
