.PHONY: dev down build mysql
dev:
	docker-compose up -d db
	docker-compose up dev
down:
	docker-compose down --rmi all
build:
	docker-compose up -d db
	docker-compose up --build -d app
mysql:
	docker-compose exec db /bin/sh -c 'mysql -u root -p'