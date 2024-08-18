cd /www/swarm-api/
make build
COMPOSE_PROJECT_NAME=gratheon docker-compose down

make migrate-db-prod

COMPOSE_PROJECT_NAME=gratheon docker-compose up -d --build