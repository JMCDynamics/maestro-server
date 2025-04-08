package main

import (
	"github.com/ardanlabs/conf/v3"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"

	"github.com/JMCDynamics/maestro-server/internal/adapters"
	"github.com/JMCDynamics/maestro-server/internal/config"
	"github.com/JMCDynamics/maestro-server/internal/dtos"
	"github.com/JMCDynamics/maestro-server/internal/server"
	"github.com/JMCDynamics/maestro-server/internal/services"
	usecases "github.com/JMCDynamics/maestro-server/internal/use-cases"
)

func main() {
	var env config.Env
	if _, err := conf.Parse("", &env); err != nil {
		panic(err)
	}

	databaseConfig := config.NewDatabaseConfig()

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("operatingsystem", dtos.ValidateOperatingSystem)
	}

	databaseGateway, err := adapters.NewDatabaseGateway(databaseConfig.UrlConnection())
	if err != nil {
		panic(err)
	}

	if err := databaseGateway.RunMigrations(); err != nil {
		panic(err)
	}

	vpnGateway := adapters.NewWireguardAdapter(env.WireguardEndpoint)

	nodeStatusService := services.NewNodeStatusService()

	cacheGateway := adapters.NewRedisCacheAdapter(databaseConfig.RedisUrlConnection(), databaseConfig.RedisPassword, 0)

	go func() {
		for key := range cacheGateway.ListenExpiredKeys() {
			log.Info().Str("node-id", key).Msg("node expired")

			nodeStatusService.SetStatus(dtos.NodeStatus{
				Id:     key,
				Status: dtos.DOWN,
			})
		}
	}()

	findNodesUseCase := usecases.NewLoggerUseCase(
		usecases.NewFindNodesUseCase(databaseGateway, cacheGateway),
	)
	findNodeUseCase := usecases.NewLoggerUseCase(
		usecases.NewFindNodeUseCase(databaseGateway, cacheGateway),
	)
	createNodeUseCase := usecases.NewLoggerUseCase(
		usecases.NewCreateNode(databaseGateway, vpnGateway),
	)
	authenticateUserUseCase := usecases.NewAuthenticateUserUseCase(
		databaseGateway,
		env.MaestroSecretKey,
	)
	setUpNodeUseCase := usecases.NewSetNodeUpUseCase(cacheGateway)
	updateNodeUseCase := usecases.NewUpdateNodeUseCase(databaseGateway, cacheGateway)

	createDefaultUser := usecases.NewCreateDefaultUserUseCase(databaseGateway, vpnGateway)

	defaultUser := env.DefaultUser()
	response, err := createDefaultUser.Execute(defaultUser)
	if err != nil {
		panic(err)
	}

	if !response.AlreadyExists {
		log.Info().
			Str("username", defaultUser.Username).
			Str("password", defaultUser.Password).
			Str("vpn-config", response.VpnConfig).
			Msg("default user created")
	}

	maestro := server.NewMaestroServer(
		env,
		findNodesUseCase,
		createNodeUseCase,
		findNodeUseCase,
		authenticateUserUseCase,
		setUpNodeUseCase,
		nodeStatusService,
		updateNodeUseCase,
	)
	if err := maestro.Run(); err != nil {
		panic(err)
	}
}
