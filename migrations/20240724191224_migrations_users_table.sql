-- +goose Up
create table users (
   id serial primary key,
   name varchar(100) not null,
   email varchar(100) not null unique,
   password varchar(64) not null,
   role smallint not null,
   status int not null,
   created_at timestamp not null default now(),
   updated_at timestamp
);

-- +goose Down
drop table chats;
