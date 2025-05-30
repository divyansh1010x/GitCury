package main

import (
	"github.com/lakshyajain-0291/gitcury/cmd"
	"github.com/lakshyajain-0291/gitcury/utils"
	"fmt"
	"os"
)

// Version information - these will be set during build by GoReleaser
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// Old server code left for reference
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
	// Set version information for use in commands
	cmd.SetVersionInfo(version, commit, date)

	// Direct version flag check for simple usage
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("GitCury %s (commit %s, built on %s)\n", version, commit, date)
		os.Exit(0)
	}

	utils.Debug(fmt.Sprintf("Starting GitCury %s", version))
	cmd.Execute()
}
