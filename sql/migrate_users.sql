alter table users change column username id varchar(30);
alter table users drop column password;
alter table users drop column userid;
alter table users drop column userlevel;
alter table users drop column timestamp;
alter table users drop column instapaper_username;
alter table users drop column instapaper_password;
alter table users drop column token;
