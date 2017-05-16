package main

import (
	"fmt"
	"github.com/ratorx/chumenu-go/api"
	// "log"
	"net/http"
)

const key = "INSERT_KEY_HERE"
const cert = "INSERT CERT HERE"

func webhook(w http.ResponseWriter, request *http.Request) {
	if request.Method == "GET" {

		url_queries := request.URL.Query()
		switch {
		case url_queries.Get("hub.mode") != "subscribe":
			fmt.Println("hub.mode is not subscribe")
			http.Error(w, "hub.mode is not subscribe", http.StatusBadRequest)
		case url_queries.Get("hub.verify_token") != "9ae4543111":
			fmt.Println("Invalid hub.verify_token")
			http.Error(w, "Invalid hub.verify_token", http.StatusBadRequest)
		default:
			fmt.Println("OK")
			fmt.Fprintf(w, url_queries.Get("hub.challenge"))
			return
		}
	} else if request.Method == "POST" {
		// Implement message commands (possibly with goroutines?)
	}
}

func main() {
	fmt.Println("Starting")
	http.HandleFunc("/webhook", webhook)                   // set router
	err := http.ListenAndServeTLS(":4999", cert, key, nil) // set listen port
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
