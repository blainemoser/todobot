-- add your UP SQL here
[STATEMENT] CREATE TABLE events (
    id INT(6) UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id INT(6) UNSIGNED NULL,
    CONSTRAINT events_user_id_foreign FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE set null,
    etext TEXT,
    channel VARCHAR(255) NULL,
    ts FLOAT,
    etype VARCHAR(255) NULL,
    schedule INT(10) NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- [DIRECTION] -- do not alter this line!
-- add your DOWN SQL here

[STATEMENT] drop table events;