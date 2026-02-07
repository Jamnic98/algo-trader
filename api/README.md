# Build docker container

```sh
docker build -t algo-trader-api:1.0 -f ./Dockerfile .
```

# Run docker container

```sh
docker run -p 8080:8080 --env-file .env.local --add-host=host.docker.internal:host-gateway algo-trader-api:1.0
```
