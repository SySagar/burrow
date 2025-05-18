// file: webserver.go
package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello from your local web server!")
	})

	fmt.Println("Web server running on http://localhost:5000")
	http.ListenAndServe(":5000", nil)
}
