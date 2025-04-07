package main

import (
	"GitCury/cmd"
)

// func main() {
// 	router := mux.NewRouter()

// 	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
// 		fmt.Fprintln(w, "Welcome to GitCury!")
// 	})

// 	router.HandleFunc("/config", handlers.ConfigHandler).Methods("GET")
// 	router.HandleFunc("/config", handlers.ConfigHandler).Methods("POST")
// 	router.HandleFunc("/getallmsgs", handlers.PrepareCommitMessagesHandler).Methods("GET")
// 	router.HandleFunc("/getonemsgs", handlers.PrepareCommitMessagesOne).Methods("GET")
// 	router.HandleFunc("/commitall", handlers.CommitAllFiles).Methods("GET")
// 	router.HandleFunc("/commitone", handlers.CommitFolder).Methods("GET")
// 	router.HandleFunc("/pushall", handlers.PushAll).Methods("GET")
// 	router.HandleFunc("/pushone", handlers.PushOne).Methods("GET")

// 	err := http.ListenAndServe(":8080", router)
// 	if err != nil {
// 		utils.Error("Error starting server: " + err.Error())
// 		return
// 	}

// 	utils.Info("Starting server on :8080")
// }

func main() {
	cmd.Execute()
}
