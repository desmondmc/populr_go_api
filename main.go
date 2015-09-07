package main

import (
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

	if !db.HasTable(&User{}) {
		db.CreateTable(&User{})
	}

	if err != nil {
		panic(err)
	}
	defer db.Close()

	appC := appContext{&db}

	commonHandlers := alice.New(context.ClearHandler, loggingHandler, recoverHandler, acceptHandler)
	router := NewRouter()
	router.Get("/user/:id", commonHandlers.ThenFunc(appC.getUserHandler))
	router.Put("/updateuser/:id", commonHandlers.Append(contentTypeHandler, bodyHandler(UserResource{})).ThenFunc(appC.updateUserHandler))
	router.Delete("/user/:id", commonHandlers.ThenFunc(appC.deleteUserHandler))
	router.Get("/users", commonHandlers.ThenFunc(appC.getUsersHandler))
	router.Post("/signup", commonHandlers.Append(contentTypeHandler, bodyHandler(UserResource{})).ThenFunc(appC.createUserHandler))
	http.ListenAndServe(":8080", router)
}
