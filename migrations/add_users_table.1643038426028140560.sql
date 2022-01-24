-- add your UP SQL here

[STATEMENT] CREATE TABLE users (
	id INT(6) UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    uhash VARCHAR(1000) NOT NULL,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- [DIRECTION] -- do not alter this line!
-- add your DOWN SQL here

[STATEMENT] drop table users;

