package server

import (
	"net/http"
	"time"

	"github.com/JMCDynamics/maestro-server/internal/config"
	"github.com/JMCDynamics/maestro-server/internal/dtos"
	"github.com/JMCDynamics/maestro-server/internal/handlers"
	"github.com/JMCDynamics/maestro-server/internal/interfaces"
	"github.com/JMCDynamics/maestro-server/internal/middlewares"
	"github.com/JMCDynamics/maestro-server/internal/services"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type maestroServer struct {
	config                  config.Env
	findNodesUseCase        interfaces.IUseCase[any, []dtos.Node]
	createNodeUseCase       interfaces.IUseCase[dtos.CreateNodeDTO, dtos.Node]
	findNodeUseCase         interfaces.IUseCase[string, dtos.Node]
	authenticateUserUseCase interfaces.IUseCase[dtos.AuthUserDTO, string]
	setUpNodeUseCase        interfaces.IUseCase[string, any]
	nodeStatusService       *services.NodeStatusService
	updateNodeUseCase       interfaces.IUseCase[dtos.UpdateNodeDTO, dtos.Node]
}

func NewMaestroServer(
	config config.Env,
	findNodesUseCase interfaces.IUseCase[any, []dtos.Node],
	createNodeUseCase interfaces.IUseCase[dtos.CreateNodeDTO, dtos.Node],
	findNodeUseCase interfaces.IUseCase[string, dtos.Node],
	authenticateUserUseCase interfaces.IUseCase[dtos.AuthUserDTO, string],
	setUpNodeUseCase interfaces.IUseCase[string, any],
	nodeStatusService *services.NodeStatusService,
	updateNodeUseCase interfaces.IUseCase[dtos.UpdateNodeDTO, dtos.Node],
) *maestroServer {
	return &maestroServer{
		config:                  config,
		findNodesUseCase:        findNodesUseCase,
		createNodeUseCase:       createNodeUseCase,
		findNodeUseCase:         findNodeUseCase,
		authenticateUserUseCase: authenticateUserUseCase,
		setUpNodeUseCase:        setUpNodeUseCase,
		nodeStatusService:       nodeStatusService,
		updateNodeUseCase:       updateNodeUseCase,
	}
}

func (s *maestroServer) Run() error {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "PATCH", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	authMiddleware := middlewares.NewAuthMiddleware(s.config.MaestroSecretKey)

	nodeHandler := handlers.NewNodeHandler(
		s.findNodesUseCase,
		s.createNodeUseCase,
		s.findNodeUseCase,
		s.setUpNodeUseCase,
		s.nodeStatusService,
		s.updateNodeUseCase,
	)

	authHandler := handlers.NewAuthHandler(s.authenticateUserUseCase)
	r.POST("/auth", authHandler.HandleAuth)

	nodeGroups := r.Group("/nodes")
	{
		nodeGroups.GET("/events", nodeHandler.HandleListenNodesStatus)
		nodeGroups.PATCH(":id", nodeHandler.HandleUpdateStatusNode)
		nodeGroups.PUT(":id", nodeHandler.HandleUpdateNode)

		nodeGroups.Use(authMiddleware.AuthMiddleware())
		nodeGroups.GET("", nodeHandler.HandleGetNodes)
		nodeGroups.POST("", nodeHandler.HandleCreateNode)
		nodeGroups.GET(":id", nodeHandler.HandleGetNode)
		nodeGroups.GET(":id/proxy-sse", nodeHandler.HandleNodeProxySSE)
		nodeGroups.Any(":id/proxy", nodeHandler.HandleNodeProxy)
	}

	r.POST("/logout", authMiddleware.AuthMiddleware(), authHandler.HandleLogout)
	r.GET("/me", authMiddleware.AuthMiddleware(), func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, "is authenticated")
	})

	return r.Run(":6276")
}
