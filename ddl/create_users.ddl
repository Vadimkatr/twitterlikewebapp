CREATE TABLE users ( 
    account_id INT NOT NULL AUTO_INCREMENT, 
    email VARCHAR(255) NOT NULL,
    username VARCHAR(255) NOT NULL,
    encrypted_password VARCHAR(255) NOT NULL,
    UNIQUE(account_id),
    UNIQUE(email),
    PRIMARY KEY(account_id)
);
