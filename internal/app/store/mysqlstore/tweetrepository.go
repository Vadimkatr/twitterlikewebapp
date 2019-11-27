package mysqlstore

import "github.com/Vadimkatr/twitterlikewebapp/internal/app/model"

type TweetRepository struct {
	store *Store
}

func (r *TweetRepository) Create(t *model.Tweet) error {
	if err := t.Validate(); err != nil {
		return err
	}

	row, err := r.store.db.Query(
		"INSERT INTO tweets (message, user_id) VALUES (?, ?)",
		t.Message,
		t.UserId,
	)

	defer row.Close()
	if err != nil {
		return err
	}
	for row.Next() {
		err := row.Scan(&t.Id, &t.PostTime)
		return err
	}
	return nil
}
