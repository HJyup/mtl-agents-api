package handler

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	pb "github.com/HJyup/mtl-common/api"
	"github.com/HJyup/mtl-common/utils"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"google.golang.org/grpc"
)

type AgentGateway interface {
	AgentWebsocketStream(ctx context.Context, opts ...grpc.CallOption) (pb.AgentService_AgentWebsocketStreamClient, error)
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
	agentRouter.Handle("/ws", utils.TokenAuthMiddleware(http.HandlerFunc(h.HandleWebsocket))).Methods("GET")
}

type WebSocketMessage struct {
	Type     string            `json:"type"`
	Content  string            `json:"content"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

func (h *AgentHandler) HandleWebsocket(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Hour)
	defer cancel()

	stream, err := h.gateway.AgentWebsocketStream(ctx)
	if err != nil {
		errorMsg := WebSocketMessage{
			Type:    "ERROR",
			Content: fmt.Sprintf("Failed to create agent stream: %v", err),
		}
		conn.WriteJSON(errorMsg)
		return
	}

	initMsg := &pb.AgentMessage{
		Type:    pb.MessageType_INITIALIZE,
		UserId:  userID,
		Content: "initial content",
	}
	if err = stream.Send(initMsg); err != nil {
		errorMsg := WebSocketMessage{
			Type:    "ERROR",
			Content: fmt.Sprintf("Failed to initialize agent: %v", err),
		}
		conn.WriteJSON(errorMsg)
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)

	done := make(chan struct{})
	var once sync.Once
	closeDone := func() {
		once.Do(func() {
			close(done)
		})
	}

	go func() {
		defer wg.Done()
		for {
			select {
			case <-done:
				return
			default:
				response, err := stream.Recv()
				if err == io.EOF {
					closeDone()
					return
				}
				if err != nil {
					log.Printf("Error receiving from stream: %v", err)
					errorMsg := WebSocketMessage{
						Type:    "ERROR",
						Content: fmt.Sprintf("Stream error: %v", err),
					}
					conn.WriteJSON(errorMsg)
					closeDone()
					return
				}

				wsMsg := WebSocketMessage{
					Content:  response.Content,
					Metadata: response.Metadata,
				}

				switch response.Type {
				case pb.MessageType_AGENT_RESPONSE:
					wsMsg.Type = "AGENT_RESPONSE"
				case pb.MessageType_ERROR:
					wsMsg.Type = "ERROR"
				case pb.MessageType_CLOSE:
					wsMsg.Type = "CLOSE"
					closeDone()
				default:
					wsMsg.Type = "UNKNOWN"
				}

				if err := conn.WriteJSON(wsMsg); err != nil {
					log.Printf("Error writing to websocket: %v", err)
					closeDone()
					return
				}
			}
		}
	}()

	go func() {
		defer wg.Done()
		for {
			select {
			case <-done:
				return
			default:
				var wsMsg WebSocketMessage
				if err := conn.ReadJSON(&wsMsg); err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway,
						websocket.CloseAbnormalClosure) {
						log.Printf("WebSocket closed unexpectedly: %v", err)
					}
					closeMsg := &pb.AgentMessage{
						Type:    pb.MessageType_CLOSE,
						UserId:  userID,
						Content: "Connection closed by client",
					}
					if err = stream.Send(closeMsg); err != nil {
						log.Printf("Error sending close message: %v", err)
					}
					closeDone()
					return
				}

				var msgType pb.MessageType
				switch wsMsg.Type {
				case "USER_MESSAGE":
					msgType = pb.MessageType_USER_MESSAGE
				case "CLOSE":
					msgType = pb.MessageType_CLOSE
					closeDone()
				default:
					log.Printf("Unknown message type: %s", wsMsg.Type)
					continue
				}

				grpcMsg := &pb.AgentMessage{
					Type:     msgType,
					UserId:   userID,
					Content:  wsMsg.Content,
					Metadata: wsMsg.Metadata,
				}

				if err = stream.Send(grpcMsg); err != nil {
					log.Printf("Error sending to stream: %v", err)
					errorMsg := WebSocketMessage{
						Type:    "ERROR",
						Content: fmt.Sprintf("Failed to send message: %v", err),
					}
					conn.WriteJSON(errorMsg)
					closeDone()
					return
				}
			}
		}
	}()

	wg.Wait()
}
