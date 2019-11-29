CREATE TABLE users ( 
    id INT NOT NULL AUTO_INCREMENT, 
    email VARCHAR(255) NOT NULL,
    username VARCHAR(255) NOT NULL,
    encrypted_password VARCHAR(255) NOT NULL,
    UNIQUE(username),
    UNIQUE(email),
    PRIMARY KEY(id)
);

CREATE TABLE tweets (
    id INT NOT NULL AUTO_INCREMENT,
    message VARCHAR(255) NOT NULL, 
    post_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, 
    user_id INT, 
    PRIMARY KEY (id), 
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE subscribers (
    id INT NOT NULL AUTO_INCREMENT,
    user_id INT, 
    publisher_user_id INT,
    PRIMARY KEY (id), 
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (publisher_user_id) REFERENCES users(id)
);
