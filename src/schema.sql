drop database pastemin;
create database pastemin;
\c pastemin;

create table pastes (
    id          varchar(8) unique,
    paste       text
);
