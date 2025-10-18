include .env
export

.PHONY: refresh

refresh:
	git pull origin master
	docker compose build app
	docker compose up -d db
	until docker compose exec db pg_isready -U $${POSTGRES_USER}; do sleep 1; done
	sleep 2
	docker compose up -d app
