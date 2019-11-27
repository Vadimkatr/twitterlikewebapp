package mysqlstore

import "github.com/Vadimkatr/twitterlikewebapp/internal/app/model"

type SubscriberRepository struct {
	store *Store
}

func (r *SubscriberRepository) Create(s *model.Subscriber) error {
	row, err := r.store.db.Query(
		"INSERT INTO subscribers (user_id, publisher_user_id) VALUES (?, ?)",
		s.UserId,
		s.PublisherUserId,
	)

	defer row.Close()
	if err != nil {
		return err
	}

	for row.Next() {
		err := row.Scan(&s.Id)
		return err
	}
	
	return nil
}
