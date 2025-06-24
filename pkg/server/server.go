package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
)

func Run() error {
	port := 7540
	if p := os.Getenv("TODO_PORT"); p != "" {
		if portNum, err := strconv.Atoi(p); err == nil {
			port = portNum
		}
	}

	fs := http.FileServer(http.Dir("web"))
	http.Handle("/", fs)

	fmt.Printf("Server listening on :%d\n", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
