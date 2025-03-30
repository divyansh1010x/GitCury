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

	router.HandleFunc("/config", handlers.ConfigHandler).Methods("POST");
	router.HandleFunc("/getmessages",handlers.PrepareCommitMessagesHandler).Methods("GET");
	router.HandleFunc("/commit",handlers.CommitAllFiles).Methods("GET");
	router.HandleFunc("/commit",handlers.CommitFolder).Methods("POST");
	
	utils.Info("Starting server on :8080")
	http.ListenAndServe(":8080", router)
}
