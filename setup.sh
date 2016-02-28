#!/bin/bash

cp sample_config.json config.json
go get -u github.com/jackc/pgx
go get -u github.com/valyala/fasthttp
go get -u github.com/buaazp/fasthttprouter
