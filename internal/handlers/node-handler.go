package handlers

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"maps"

	"github.com/JMCDynamics/maestro-server/internal/dtos"
	"github.com/JMCDynamics/maestro-server/internal/interfaces"
	"github.com/JMCDynamics/maestro-server/internal/services"
	usecases "github.com/JMCDynamics/maestro-server/internal/use-cases"
	"github.com/gin-gonic/gin"
)

type nodeHandler struct {
	findNodesUseCase  interfaces.IUseCase[any, []dtos.Node]
	findNodeUseCase   interfaces.IUseCase[string, dtos.Node]
	createNodeUseCase interfaces.IUseCase[dtos.CreateNodeDTO, dtos.Node]
	setNodeUpUseCase  interfaces.IUseCase[string, any]
	nodeStatusService *services.NodeStatusService
	updateNodeUseCase interfaces.IUseCase[dtos.UpdateNodeDTO, dtos.Node]
}

func NewNodeHandler(
	findNodesUseCase interfaces.IUseCase[any, []dtos.Node],
	createNodeUseCase interfaces.IUseCase[dtos.CreateNodeDTO, dtos.Node],
	findNodeUseCase interfaces.IUseCase[string, dtos.Node],
	setNodeUpUseCase interfaces.IUseCase[string, any],
	nodeStatusService *services.NodeStatusService,
	updateNodeUseCase interfaces.IUseCase[dtos.UpdateNodeDTO, dtos.Node],
) nodeHandler {
	return nodeHandler{
		findNodesUseCase:  findNodesUseCase,
		createNodeUseCase: createNodeUseCase,
		findNodeUseCase:   findNodeUseCase,
		setNodeUpUseCase:  setNodeUpUseCase,
		nodeStatusService: nodeStatusService,
		updateNodeUseCase: updateNodeUseCase,
	}
}

func (h *nodeHandler) HandleGetNodes(c *gin.Context) {
	nodes, err := h.findNodesUseCase.Execute(nil)
	if err != nil {
		return
	}

	response := dtos.NewDefaultResponse("action exectued with success", nodes)
	c.JSON(http.StatusOK, response)
}

func (h *nodeHandler) HandleGetNode(c *gin.Context) {
	nodeId := c.Param("id")
	nodes, err := h.findNodeUseCase.Execute(nodeId)
	if err != nil {
		return
	}

	response := dtos.NewDefaultResponse("action exectued with success", nodes)
	c.JSON(http.StatusOK, response)
}

func (h *nodeHandler) HandleCreateNode(c *gin.Context) {
	nodes, err := h.findNodesUseCase.Execute(nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dtos.NewDefaultResponse("failed to find nodes", nil))
		return
	}

	if len(nodes) >= 4 {
		c.JSON(http.StatusForbidden, dtos.NewDefaultResponse("maximum number of nodes reached", nil))
		return
	}

	var data dtos.CreateNodeDTO
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, dtos.NewDefaultResponse(err.Error(), nil))
		return
	}

	node, err := h.createNodeUseCase.Execute(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dtos.NewDefaultResponse("failed to create node", nil))
		return
	}

	response := dtos.NewDefaultResponse("action exectued with success", node)
	c.JSON(http.StatusCreated, response)
}

func (h *nodeHandler) HandleNodeProxySSE(c *gin.Context) {
	nodeId := c.Param("id")
	path := c.Query("path")
	if path == "" {
		response := dtos.NewDefaultResponse("param path is empty", nil)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	node, err := h.findNodeUseCase.Execute(nodeId)
	if err == usecases.ErrNodeNotFound {
		response := dtos.NewDefaultResponse(err.Error(), nil)
		c.JSON(http.StatusNotFound, response)
		return
	}

	if err != nil {
		response := dtos.NewDefaultResponse("unable to find node", nil)
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	sseSourceURL := fmt.Sprintf("http://%s:%s%s", node.VpnAddress, "9842", path)

	client := &http.Client{
		Timeout: 0,
	}
	req, err := http.NewRequest("GET", sseSourceURL, nil)
	if err != nil {
		c.String(http.StatusInternalServerError, "Erro ao criar requisição: %s", err)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		c.String(http.StatusInternalServerError, "Erro ao conectar ao servidor SSE: %s", err)
		return
	}
	defer resp.Body.Close()

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	done := make(chan struct{})

	go func() {
		<-ctx.Done()
		log.Println("Conexão cliente cancelada")
		cancel()
		if resp.Body != nil {
			resp.Body.Close()
		}
		done <- struct{}{}
	}()

	reader := bufio.NewReader(resp.Body)

loop:
	for {
		select {
		case <-done:
			break loop
		default:
			line, err := reader.ReadBytes('\n')
			if err != nil {
				if err != io.EOF {
					log.Printf("Error reading stream: %v", err)
				}
				return
			}

			if _, err := c.Writer.Write(line); err != nil {
				log.Printf("Error writing to client: %v", err)
				return
			}
			c.Writer.Flush()
		}
	}

	fmt.Println("finished sse proxy")
}

func (h *nodeHandler) HandleNodeProxy(c *gin.Context) {
	nodeId := c.Param("id")
	path := c.Query("path")
	if path == "" {
		response := dtos.NewDefaultResponse("param path is empty", nil)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	node, err := h.findNodeUseCase.Execute(nodeId)
	if err == usecases.ErrNodeNotFound {
		response := dtos.NewDefaultResponse(err.Error(), nil)
		c.JSON(http.StatusNotFound, response)
		return
	}

	if err != nil {
		response := dtos.NewDefaultResponse("unable to find node", nil)
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	targetURL := fmt.Sprintf("http://%s:%s%s", node.VpnAddress, "9842", path)

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest(c.Request.Method, targetURL, c.Request.Body)
	if err != nil {
		response := dtos.NewDefaultResponse("unable to create a request", nil)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		response := dtos.NewDefaultResponse("failed to connect to the node", nil)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}
	defer resp.Body.Close()

	maps.Copy(c.Writer.Header(), resp.Header)

	c.Status(resp.StatusCode)

	_, err = io.Copy(c.Writer, resp.Body)
	if err != nil {
		log.Printf("Error copying response body: %v", err)
	}
}

func (h *nodeHandler) HandleUpdateStatusNode(c *gin.Context) {
	nodeId := c.Param("id")

	node, err := h.findNodeUseCase.Execute(nodeId)
	if err == usecases.ErrNodeNotFound {
		response := dtos.NewDefaultResponse("unable to find node", nil)
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	_, err = h.setNodeUpUseCase.Execute(node.Id)
	if err != nil {
		response := dtos.NewDefaultResponse(err.Error(), nil)
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	h.nodeStatusService.SetStatus(dtos.NodeStatus{
		Id:     node.Id,
		Status: dtos.UP,
	})

	response := dtos.NewDefaultResponse("action exectued with success", nodeId)
	c.JSON(http.StatusOK, response)
}

func (h *nodeHandler) HandleListenNodesStatus(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	done := make(chan struct{})

	go func() {
		<-ctx.Done()
		log.Println("Conexão cliente cancelada")
		cancel()
		done <- struct{}{}
	}()

	for {
		select {
		case <-done:
			log.Println("Finalizando conexão devido ao cancelamento do contexto.")
			return
		case nodeStatus, ok := <-h.nodeStatusService.ListenStatus():
			if !ok {
				log.Println("Canal de status fechado")
				return
			}

			dataJson, err := json.Marshal(nodeStatus)
			if err != nil {
				log.Printf("Erro ao serializar status do nó: %s", err)
				return
			}

			fmt.Fprintf(c.Writer, "data: %s\n\n", dataJson)
			c.Writer.Flush()
		}
	}
}

func (h *nodeHandler) HandleUpdateNode(c *gin.Context) {
	nodeId := c.Param("id")

	var data dtos.UpdateNodeDTO
	if err := c.ShouldBindJSON(&data); err != nil {
		response := dtos.NewDefaultResponse(err.Error(), nil)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	data.Id = nodeId

	node, err := h.updateNodeUseCase.Execute(data)
	if err != nil {
		response := dtos.NewDefaultResponse(err.Error(), nil)
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := dtos.NewDefaultResponse("action exectued with success", node)
	c.JSON(http.StatusOK, response)
}
