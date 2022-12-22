# gratheon / swarm-api
Main monolith service to manage beehive data.


## Architecture

```mermaid
flowchart LR
    web-app("<a href='https://github.com/Gratheon/web-app'>web-app</a>") --> graphql-router
    web-app --"subscribe to events"--> event-stream-filter("<a href='https://github.com/Gratheon/event-stream-filter'>event-stream-filter</a>") --> redis
    
    graphql-router --> swarm-api("<a href='https://github.com/Gratheon/swarm-api'>swarm-api</a>") --> mysql[(mysql)]
    graphql-router --> swarm-api --> redis[("<a href='https://github.com/Gratheon/redis'>redis pub-sub</a>")]
    
    graphql-router --> graphql-schema-registry
```


## Development
Based on [gqlgen](https://gqlgen.com/getting-started/).

- re-run graphql schema -> code generator:
```
make gen
```

- start db
```
docker-compose up
``` 

- run server:
```
make run
```

## Building
```
make build
```

## Deployment
```
make deploy
```
