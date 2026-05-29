#!/bin/bash

if [ -f .env ]; then
    source .env
fi

goose -dir ./sql/migrations postgres $DATABASE_URL up
