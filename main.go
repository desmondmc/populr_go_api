package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/context"
	"github.com/jmoiron/sqlx"
	"github.com/justinas/alice"
	_ "github.com/lib/pq"
)

type appContext struct {
	db *sqlx.DB
}

func main() {
	if os.Getenv("PORT") == "" {
		log.Println("Please set PORT env variable.")
		return
	}

	portString := ":" + os.Getenv("PORT")
	dokku_db := "postgres://root:AdWNMo0dvsV0CVhK@172.17.42.1:32768/db"

	log.Println("Database String: ", dokku_db)

	db, err := sqlx.Connect("postgres", dokku_db)
	if err != nil {
		fmt.Printf("sql.Open error: %v\n", err)
		return
	}
	defer db.Close()

	log.Println("Setting up database..")
	dbSetup(db)

	appC := appContext{db}

	commonHandlers := alice.New(context.ClearHandler, loggingHandler, recoverHandler, acceptHandler)
	loggedInCommonHandlers := commonHandlers.Append(contentTypeHandler)

	log.Println("Setting up routes...")
	router := NewRouter()
	router.Get("/user/:id", commonHandlers.ThenFunc(appC.getUserHandler))
	router.Get("/followers", loggedInCommonHandlers.ThenFunc(appC.getUserFollowersHandler))
	router.Get("/following", loggedInCommonHandlers.ThenFunc(appC.getUsersFollowingHandler))
	router.Get("/users", commonHandlers.ThenFunc(appC.getUsersHandler))
	router.Get("/searchusers/:term", loggedInCommonHandlers.ThenFunc(appC.searchUsersHandler))
	router.Get("/messages", loggedInCommonHandlers.ThenFunc(appC.getMessagesHandler))
	router.Post("/signup", commonHandlers.Append(contentTypeHandler, bodyHandler(UserResource{})).ThenFunc(appC.createUserHandler))
	router.Post("/follow/:id", loggedInCommonHandlers.ThenFunc(appC.followUserHandler))
	router.Post("/readmessage/:id", loggedInCommonHandlers.ThenFunc(appC.readMessageHandler))
	router.Post("/message", commonHandlers.Append(contentTypeHandler, bodyHandler(RecieveMessageResource{})).ThenFunc(appC.postMessageHandler))
	router.Delete("/unfollow/:id", loggedInCommonHandlers.ThenFunc(appC.unfollowUserHandler))

	log.Println("Listening...")
	http.ListenAndServe(portString, router)
}

var schema = `
CREATE TABLE users (
	id SERIAL NOT NULL PRIMARY KEY,
    username text,
    password text
);

CREATE TABLE user_followers (
      user_id    int REFERENCES users (id) ON UPDATE CASCADE ON DELETE CASCADE, 
      follower_id int REFERENCES users (id) ON UPDATE CASCADE ON DELETE CASCADE
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

func dbSetup(db *sqlx.DB) {
	db.Exec(schema)
}
