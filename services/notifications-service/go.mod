module github.com/lucas/notifications-service

go 1.22.2

require (
	github.com/lucas/shared v0.0.0
	github.com/segmentio/kafka-go v0.4.48
)

replace github.com/lucas/shared => ../../shared

require (
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	github.com/stretchr/testify v1.9.0 // indirect
	golang.org/x/net v0.25.0 // indirect
)
