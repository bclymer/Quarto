package admin

import (
	"fmt"
    "net/http"
    "quarto/realtime"
    "log"
)

func Index(w http.ResponseWriter, r *http.Request) {
	names := realtime.GetUserMap()
	fmt.Fprint(w, names)
	log.Print("Done")
}