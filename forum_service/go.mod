module github.com/kprf42/dolgova/forum_service

go 1.24.2

require (
	github.com/go-chi/chi/v5 v5.2.1
	github.com/google/uuid v1.6.0
	github.com/kprf42/dolgova/pkg/logger v0.0.0-00010101000000-000000000000
	github.com/kprf42/dolgova/proto v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.72.1
)

require (
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
)

require (
	github.com/golang-jwt/jwt/v5 v5.2.2
	github.com/golang-migrate/migrate/v4 v4.18.3
	github.com/gorilla/websocket v1.5.3
	github.com/joho/godotenv v1.5.1
	github.com/mattn/go-sqlite3 v1.14.28
	golang.org/x/net v0.40.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.25.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250218202821-56aae31c358a // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)

replace github.com/kprf42/dolgova/proto => ../proto

replace github.com/kprf42/dolgova/pkg/logger => ../pkg/logger
