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

func (r *TweetRepository) FindTweetsFromSubscriptions(id int) ([]string, error) {
	rows, err := r.store.db.Query(
		"SELECT message FROM tweets WHERE user_id IN" +
		"	(SELECT publisher_user_id FROM subscribers WHERE user_id = ?)" +
		"	ORDER BY post_time DESC",
		id,
	)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	var tweets []string
    for rows.Next() {
		var tweet string
        err := rows.Scan(&tweet)
        if err != nil {
			return nil, err
        }
        tweets = append(tweets, tweet)
    }
    if err := rows.Err(); err != nil {
        return nil, err
	}
	return tweets, nil
}

func (r *TweetRepository) GetAllUserTweets(userId int) ([]string, error) {

	rows, err := r.store.db.Query(
		"SEKECT message FROM tweets WHERE user_id = ?",
		userId,
	)

	defer rows.Close()
	if err != nil {
		return nil, err
	}
	var tweets []string
	var tweet string
    for rows.Next() {
        err := rows.Scan(&tweet)
        if err != nil {
			return nil, err
        }
        tweets = append(tweets, tweet)
    }
    if err := rows.Err(); err != nil {
        return nil, err
	}
	return tweets, nil
}
