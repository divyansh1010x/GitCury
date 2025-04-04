package main

import (
	"GitCury/handlers"
	"GitCury/utils"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Welcome to GitCury!")
	})

	router.HandleFunc("/config", handlers.ConfigHandler).Methods("GET")
	router.HandleFunc("/config", handlers.ConfigHandler).Methods("POST")
	router.HandleFunc("/getallmsgs", handlers.PrepareCommitMessagesHandler).Methods("GET")
	router.HandleFunc("/getonemsgs", handlers.PrepareCommitMessagesOne).Methods("Get")
	router.HandleFunc("/commitall", handlers.CommitAllFiles).Methods("GET")
	router.HandleFunc("/commitone", handlers.CommitFolder).Methods("Get")
	router.HandleFunc("/pushall", handlers.PushAll).Methods("POST")
	router.HandleFunc("/pushfolder", handlers.PushOne).Methods("POST")

	err := http.ListenAndServe(":8080", router)
	if err != nil {
		utils.Error("Error starting server: " + err.Error())
		return
	}

	utils.Info("Starting server on :8080")
}
