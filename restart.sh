cd /www/swarm-api/
make build
COMPOSE_PROJECT_NAME=gratheon docker-compose down
COMPOSE_PROJECT_NAME=gratheon docker-compose up -d --build