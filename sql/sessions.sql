drop table if exists `sessions`;
create table sessions (
	user_id varchar(255) NOT NULL,
	session_hash char(255) NOT NULL PRIMARY KEY
);
