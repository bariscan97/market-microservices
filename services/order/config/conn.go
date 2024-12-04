package config

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	Host                  string
	Port                  string
	User                  string
	Password              string
	DbName                string
	MaxConnections        string
	MaxConnectionIdleTime string
}

func GetConnectionPool(config Config) *pgxpool.Pool {
	connString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable pool_max_conns=%s pool_max_conn_idle_time=%s",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.DbName,
		config.MaxConnections,
		config.MaxConnectionIdleTime)

	connConfig, parseConfigErr := pgxpool.ParseConfig(connString)
	if parseConfigErr != nil {
		panic(parseConfigErr)
	}
	
	conn, err := pgxpool.NewWithConfig(context.Background(), connConfig)
	
	if err != nil {
		panic(err)
	}

	return conn
}