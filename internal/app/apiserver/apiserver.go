package apiserver

import (
	"net/http"
	"github.com/gorilla/mux"
)

func Start(config *Config) error {
	r := mux.NewRouter()
	r.HandleFunc("/register", handleUserCreate()).Methods("POST")
	return http.ListenAndServe(":8080", r)
}

func handleUserCreate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello"))
	}
}
