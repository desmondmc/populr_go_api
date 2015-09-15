package main

import (
	"log"
	"net/http"

	"github.com/gorilla/context"
	"github.com/jinzhu/gorm"
	"github.com/justinas/alice"
	_ "github.com/lib/pq"
)

type appContext struct {
	db *gorm.DB
}

func main() {
	db, err := gorm.Open("postgres", "user=desmondmcnamee dbname=populr sslmode=disable")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	log.Println("Setting up database..")
	dbSetup(&db)

	appC := appContext{&db}

	commonHandlers := alice.New(context.ClearHandler, loggingHandler, recoverHandler, acceptHandler)
	loggedInCommonHandlers := commonHandlers.Append(contentTypeHandler)

	log.Println("Setting up routes...")
	router := NewRouter()
	router.Get("/user/:id", commonHandlers.ThenFunc(appC.getUserHandler))
	router.Get("/followers", loggedInCommonHandlers.ThenFunc(appC.getUserFollowersHandler))
	router.Get("/users", commonHandlers.ThenFunc(appC.getUsersHandler))
	router.Post("/signup", commonHandlers.Append(contentTypeHandler, bodyHandler(UserResource{})).ThenFunc(appC.createUserHandler))
	router.Post("/follow/:userid", loggedInCommonHandlers.ThenFunc(appC.followUserHandler))

	log.Println("Listening...")
	http.ListenAndServe(":8080", router)
}

func dbSetup(db *gorm.DB) {
	db.AutoMigrate(&User{}) // Creates a table if it doesn't exist; updates the table if the struct changes.
}
