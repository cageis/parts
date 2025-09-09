CREATE TABLE products (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,
    price DECIMAL(10,2),
    user_id INT,
    FOREIGN KEY (user_id) REFERENCES users(id)
);