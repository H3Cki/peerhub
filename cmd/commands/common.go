package commands

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/H3Cki/peerhub"
)

func AnsweringsHandler(h *peerhub.Hub) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		aps, err := h.GetAnsweringPeersPrevies()
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		b, err := json.Marshal(aps)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if _, err := w.Write(b); err != nil {
			fmt.Println(err)
		}
	}
}
