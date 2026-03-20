module github.com/jakottelaar/relay-microservices/services/guilds

go 1.26.0

require (
	github.com/jakottelaar/relay-microservices/shared v0.0.0-00010101000000-000000000000
	github.com/joho/godotenv v1.5.1
	go.uber.org/zap v1.27.1
)

require go.uber.org/multierr v1.10.0 // indirect

replace github.com/jakottelaar/relay-microservices/shared => ../../shared
