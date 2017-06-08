package main

import (
	"net/http"
	"fmt"
)

func main() {
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)

	http.HandleFunc("/startListen", configHandler)
	http.ListenAndServe(":9090", nil)
}

func configHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("This is working")
	fmt.Fprintf(w, "Hello World")
	listenScoket()
}