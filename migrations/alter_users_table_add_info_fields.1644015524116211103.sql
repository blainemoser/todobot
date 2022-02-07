-- add your UP SQL here

[STATEMENT] ALTER TABLE users ADD tz VARCHAR(255) AFTER uhash; 
[STATEMENT] ALTER TABLE users ADD tz_label VARCHAR(255) AFTER tz;
[STATEMENT] ALTER TABLE users ADD tz_offset INT AFTER tz_label;

-- [DIRECTION] -- do not alter this line!
-- add your DOWN SQL here

[STATEMENT] ALTER TABLE users DROP COLUMN tz;
[STATEMENT] ALTER TABLE users DROP COLUMN tz_label;
[STATEMENT] ALTER TABLE users DROP COLUMN tz_offset;