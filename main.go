package main

import (
	"net/http"

	"github.com/gorilla/context"
	"github.com/justinas/alice"
	"gopkg.in/mgo.v2"
)

type appContext struct {
	db *mgo.Database
}

func main() {
	session, err := mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	appC := appContext{session.DB("test")}
	commonHandlers := alice.New(context.ClearHandler, loggingHandler, recoverHandler, acceptHandler)
	router := NewRouter()
	router.Get("/user/:id", commonHandlers.ThenFunc(appC.getUserHandler))
	router.Put("/updateuser/:id", commonHandlers.Append(contentTypeHandler, bodyHandler(UserResource{})).ThenFunc(appC.updateUserHandler))
	router.Delete("/user/:id", commonHandlers.ThenFunc(appC.deleteUserHandler))
	router.Get("/users", commonHandlers.ThenFunc(appC.getUsersHandler))
	router.Post("/signup", commonHandlers.Append(contentTypeHandler, bodyHandler(UserResource{})).ThenFunc(appC.createUserHandler))
	http.ListenAndServe(":8080", router)
}
