include .env
export

.PHONY: refresh build up down logs build-front

refresh:
	git pull origin master
	docker compose build app
	docker compose up -d db
	until docker compose exec db pg_isready -U $${POSTGRES_USER}; do sleep 1; done
	sleep 2
	docker compose up -d app
	docker compose logs -f app

build-front:
	cd .. && \
	cd make_front && \
	npm install && \
	npm run build && \
	rm -rf ../makeziper/front-dist/* && \
	cp -r dist/* ../make_ziper/front-dist/

build:
	docker compose build

up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f