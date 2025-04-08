package config

import (
	"fmt"

	"github.com/JMCDynamics/maestro-server/internal/dtos"
)

type Database struct {
	DatabaseHost     string
	DatabasePort     string
	DatabaseUser     string
	DatabasePassword string
	DatabaseName     string

	RedisHost     string
	RedisPort     string
	RedisPassword string
}

func NewDatabaseConfig() Database {
	return Database{
		DatabaseHost:     "localhost",
		DatabasePort:     "5432",
		DatabaseUser:     "postgres",
		DatabasePassword: "Docker",
		DatabaseName:     "maestro_db",

		RedisHost:     "localhost",
		RedisPort:     "6379",
		RedisPassword: "",
	}
}

func (d *Database) UrlConnection() string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable", d.DatabaseUser, d.DatabasePassword, d.DatabaseHost, d.DatabasePort, d.DatabaseName)
}

func (d *Database) RedisUrlConnection() string {
	return fmt.Sprintf("%s:%s", d.RedisHost, d.RedisPort)
}

type Env struct {
	WireguardEndpoint string `conf:"env:WIREGUARD_ENDPOINT"`

	MaestroUsername string `conf:"env:MAESTRO_USERNAME,default:maestro"`
	MaestroPassword string `conf:"env:MAESTRO_PASSWORD,default:root"`

	MaestroSecretKey string `conf:"env:MAESTRO_SECRET_KEY,default:maestro_key_dev"`
}

func (e *Env) DefaultUser() dtos.CreateUserDTO {
	return dtos.CreateUserDTO{
		Username: e.MaestroUsername,
		Password: e.MaestroPassword,
	}
}
