# Build Examples

## Build bot image
```bash
docker build --build-arg BUILD_TARGET=bot -t movieclub-bot:latest .
```

## Build worker image
```bash
docker build --build-arg BUILD_TARGET=worker -t movieclub-worker:latest .
```

## Run bot locally
```bash
docker run --rm \
  --env-file .env \
  -v $(pwd)/data:/root/data \
  -p 2000:2000 \
  movieclub-bot:latest
```

## Run worker locally
```bash
docker run --rm \
  --env-file .env \
  -v $(pwd)/data:/root/data \
  movieclub-worker:latest
```

## Build both with docker-compose
```bash
# Development
docker-compose build

# Production
docker-compose -f docker-compose.prod.yml build

# Rebuild without cache
docker-compose build --no-cache

# Build specific service
docker-compose build bot
docker-compose build worker
```
