package main

import (
	"fmt"
	"net/http"
	"stepikGoWebServices/handlers"
)

func main() {
	addr := ":8080"
	h := handlers.NewRealWorldApi()
	fmt.Println("start server at", addr)
	http.ListenAndServe(addr, h)
}
