CREATE TABLE subscribers (
    id INT NOT NULL AUTO_INCREMENT,
    user_id INT, 
    publisher_user_id INT,
    PRIMARY KEY (id), 
    FOREIGN KEY (user_id) REFERENCES users(account_id),
    FOREIGN KEY (publisher_user_id) REFERENCES users(account_id)
);
