SHELL := /bin/sh

-include .env
export

COMPOSE := docker compose

.DEFAULT_GOAL := up

.PHONY: up rebuild down db-init test test-cover

up:
	$(COMPOSE) up -d

rebuild:
	$(COMPOSE) up -d --build

down:
	$(COMPOSE) down

db-init:
	$(COMPOSE) up -d postgres
	$(COMPOSE) exec -T postgres psql -U "$(POSTGRES_USER)" -d "$(POSTGRES_DB)" -f /docker-entrypoint-initdb.d/init.sql

test:
	go test -v ./...

test-cover:
	go test -v -cover ./...
