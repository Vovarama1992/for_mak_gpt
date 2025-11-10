.PHONY: refresh build up down logs build-front

refresh:
	git pull origin master
	docker compose build app
	docker compose stop app
	docker compose up -d --no-deps app
	docker compose logs -f app

build-front:
	cd .. && \
	cd make_front && \
	npm install && \
	npm run build && \
	rm -rf ../repetitor/front-dist/* && \
    cp -r dist/* ../repetitor/front-dist/

build:
	docker compose build

up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f