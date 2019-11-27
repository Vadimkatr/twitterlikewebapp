package apiserver

import (
	"database/sql"
	"net/http"
	"github.com/gorilla/mux"

	"github.com/Vadimkatr/twitterlikewebapp/internal/app/store/mysqlstore"
)

func Start(config *Config) error {
	db, err := newDB(config.DatabaseURL)
	if err != nil {
		return err
	}

	defer db.Close()
	store := mysqlstore.New(db)

	r := mux.NewRouter()
	r.HandleFunc("/register", handleUserCreate()).Methods("POST")
	return http.ListenAndServe(":8080", r)
}

func handleUserCreate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello"))
	}
}

// newDB func create new db connection from databaseURL
func newDB(dbURL string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dbURL)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
