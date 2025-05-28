package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DBPath    string
	HTTPPort  int
	GRPCPort  int
	JWTSecret string
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	httpPort, err := strconv.Atoi(os.Getenv("HTTP_PORT"))
	if err != nil {
		return nil, err
	}

	grpcPort, err := strconv.Atoi(os.Getenv("GRPC_PORT"))
	if err != nil {
		return nil, err
	}

	return &Config{
		DBPath:    os.Getenv("DB_PATH"),
		HTTPPort:  httpPort,
		GRPCPort:  grpcPort,
		JWTSecret: os.Getenv("JWT_SECRET"),
	}, nil
}

// package config

// import (
// 	"os"
// 	"strconv"

// 	"github.com/joho/godotenv"
// )

// type Config struct {
// 	DBPath    string
// 	HTTPPort  int
// 	GRPCPort  int
// 	JWTSecret string
// }

// func Load() (*Config, error) {
// 	if err := godotenv.Load(); err != nil {
// 		return nil, err
// 	}

// 	httpPort, err := strconv.Atoi(os.Getenv("HTTP_PORT"))
// 	if err != nil {
// 		return nil, err
// 	}

// 	grpcPort, err := strconv.Atoi(os.Getenv("GRPC_PORT"))
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &Config{
// 		DBPath:    os.Getenv("DB_PATH"),
// 		HTTPPort:  httpPort,
// 		GRPCPort:  grpcPort,
// 		JWTSecret: os.Getenv("JWT_SECRET"),
// 	}, nil
// }
