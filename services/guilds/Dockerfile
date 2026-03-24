FROM golang:1.26.0-alpine AS build

WORKDIR /app

RUN apk add --no-cache git

# Copy shared module so the replace directive resolves
COPY shared/ ./shared/

# Copy service module files
COPY services/guilds/ ./services/guilds/

# Copy migrations
COPY services/guilds/migrations/ ./migrations/

WORKDIR /app/services/guilds

# Disable workspace, use the replace directive in go.mod instead
RUN GOWORK=off go mod download

RUN GOWORK=off CGO_ENABLED=0 GOOS=linux go build -o /app/main ./cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=build /app/main .

COPY --from=build /app/migrations ./migrations

CMD ["./main"]