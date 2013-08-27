package admin

import (
	"log"
	"fmt"
    "net/http"
)

func Index(w http.ResponseWriter, r *http.Request) {
	log.Print("Yea")
	fmt.Fprint(w, "Test")
}