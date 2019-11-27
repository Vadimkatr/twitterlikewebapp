CREATE TABLE tweets (
    id INT NOT NULL AUTO_INCREMENT,
    message VARCHAR(255) NOT NULL, 
    post_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, 
    user_id INT, 
    PRIMARY KEY (id), 
    FOREIGN KEY (user_id) REFERENCES users(account_id)
);
