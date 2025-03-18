package probeserver

import (
	"fmt"
	"log"
	"net/http"
)

func handleGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Handle the GET request logic here
	fmt.Fprintf(w, "Received GET request!\n")
	log.Println("Received GET request!")
}

func StartProbeSever() error {
	http.HandleFunc("/", handleGet)
	fmt.Printf("Enter proble server\n")
	log.Println("Enter proble server")

	var err error
	go func() {
		err = http.ListenAndServe(":8001", nil)
		log.Println("start proble server ")
		fmt.Printf("start proble server\n")
	}()

	if err != nil {
		fmt.Printf("start proble server error %s\n", err.Error())
		log.Println("start proble server error ", err)
	}

	return err
}
