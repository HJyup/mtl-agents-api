package handler

import (
	"context"
	"encoding/json"
	"github.com/HJyup/mlt-gateway/internal/models"
	pb "github.com/HJyup/mtl-common/api"
	"github.com/HJyup/mtl-common/utils"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"io"
	"net/http"
)

type AgentGateway interface {
	CreateAgentStream(ctx context.Context, req *pb.CreateAgentStreamRequest) (pb.AgentService_CreateAgentStreamClient, error)
	SendAgentMessage(ctx context.Context, req *pb.SendAgentMessageRequest) (*pb.SendAgentMessageResponse, error)
}

type AgentHandler struct {
	gateway  AgentGateway
	upgrader websocket.Upgrader
}

func NewAgentHandler(gateway AgentGateway) *AgentHandler {
	return &AgentHandler{
		gateway: gateway,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (h *AgentHandler) RegisterRoutes(router *mux.Router) {
	agentRouter := router.PathPrefix("/api/v1/agents").Subrouter()
	agentRouter.Handle("/stream", utils.TokenAuthMiddleware(http.HandlerFunc(h.HandleCreateAgentStream))).Methods("GET")
	agentRouter.Handle("/message", utils.TokenAuthMiddleware(http.HandlerFunc(h.HandleSendAgentMessage))).Methods("POST")
}

func (h *AgentHandler) HandleCreateAgentStream(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	configID := r.URL.Query().Get("config_id")
	if configID == "" {
		utils.WriteError(w, http.StatusBadRequest, "Missing config_id parameter")
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	stream, err := h.gateway.CreateAgentStream(r.Context(), &pb.CreateAgentStreamRequest{
		UserId: userID,
	})
	if err != nil {
		errorMessage := map[string]string{"error": err.Error()}
		conn.WriteJSON(errorMessage)
		return
	}

	for {
		response, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			errorMessage := map[string]string{"error": err.Error()}
			conn.WriteJSON(errorMessage)
			break
		}

		if err := conn.WriteJSON(response); err != nil {
			break
		}
	}
}

func (h *AgentHandler) HandleSendAgentMessage(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var reqBody models.SendAgentMessageRequest

	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Failed to read request body")
		return
	}
	defer r.Body.Close()

	if err = json.Unmarshal(body, &reqBody); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if reqBody.Message == "" {
		utils.WriteError(w, http.StatusBadRequest, "Message is required")
		return
	}

	resp, err := h.gateway.SendAgentMessage(r.Context(), &pb.SendAgentMessageRequest{
		UserId:  userID,
		Message: reqBody.Message,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}
