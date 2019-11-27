CREATE TABLE users ( 
    id INT NOT NULL AUTO_INCREMENT, 
    email VARCHAR(255) NOT NULL,
    username VARCHAR(255) NOT NULL,
    encrypted_password VARCHAR(255) NOT NULL,
    UNIQUE(id),
    UNIQUE(email),
    PRIMARY KEY(id)
);
