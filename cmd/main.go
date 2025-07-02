package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	fmt.Println("Taskflow API starting...")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Taskflow API is running!\n"))
	})

	fmt.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}