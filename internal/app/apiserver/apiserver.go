package apiserver

import (
	"database/sql"
	"net/http"

	"github.com/Vadimkatr/twitterlikewebapp/internal/app/store/mysqlstore"
	_ "github.com/go-sql-driver/mysql"
)

func Start(config *Config) error {
	db, err := newDB(config.DatabaseURL)
	if err != nil {
		return err
	}

	defer db.Close()
	store := mysqlstore.New(db)
	srv := newServer(store)
	
	return http.ListenAndServe(config.BindAddr, srv)
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
