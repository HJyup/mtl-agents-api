package handler

import (
	"context"
	"encoding/json"
	"github.com/HJyup/mlt-gateway/internal/models"
	pb "github.com/HJyup/mtl-common/api"
	"github.com/HJyup/mtl-common/utils"
	"github.com/gorilla/mux"
	"io"
	"net/http"
)

type ConfigurationGateway interface {
	CreateConfiguration(context.Context, *pb.CreateConfigurationRequest) (*pb.CreateConfigurationResponse, error)
	UpdateConfiguration(context.Context, *pb.UpdateConfigurationRequest) (*pb.UpdateConfigurationResponse, error)
	GetConfiguration(context.Context, *pb.GetConfigurationRequest) (*pb.GetConfigurationResponse, error)
	DeleteConfiguration(context.Context, *pb.DeleteConfigurationRequest) (*pb.DeleteConfigurationResponse, error)
}

type ConfigurationHandler struct {
	gateway ConfigurationGateway
}

func NewConfigurationHandler(gateway ConfigurationGateway) *ConfigurationHandler {
	return &ConfigurationHandler{gateway: gateway}
}

func (h *ConfigurationHandler) RegisterRoutes(router *mux.Router) {
	configRouter := router.PathPrefix("/api/v1/configurations").Subrouter()
	configRouter.Handle("", utils.TokenAuthMiddleware(http.HandlerFunc(h.HandleCreateConfiguration))).Methods("POST")
	configRouter.Handle("", utils.TokenAuthMiddleware(http.HandlerFunc(h.HandleUpdateConfiguration))).Methods("PUT")
	configRouter.Handle("/{userId}", utils.TokenAuthMiddleware(http.HandlerFunc(h.HandleGetConfiguration))).Methods("GET")
	configRouter.Handle("/{userId}", utils.TokenAuthMiddleware(http.HandlerFunc(h.HandleDeleteConfiguration))).Methods("DELETE")
}

func (h *ConfigurationHandler) HandleCreateConfiguration(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	resp, err := h.gateway.CreateConfiguration(r.Context(), &pb.CreateConfigurationRequest{
		UserId: userID,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.WriteJSON(w, http.StatusCreated, resp)
}

func (h *ConfigurationHandler) HandleUpdateConfiguration(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var reqBody models.UpdateConfigurationRequest

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

	req := &pb.UpdateConfigurationRequest{
		UserId:    userID,
		OpenAiKey: reqBody.OpenAIKey,
	}

	if reqBody.Calendar != nil {
		req.Calendar = &pb.CalendarConfig{
			GoogleApiKey: reqBody.Calendar.GoogleAPIKey,
			Context:      reqBody.Calendar.Context,
		}
	}

	if reqBody.Things != nil {
		req.Things = &pb.ThingsConfig{
			Context: reqBody.Things.Context,
		}
	}

	resp, err := h.gateway.UpdateConfiguration(r.Context(), req)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if !resp.Success {
		utils.WriteError(w, http.StatusBadRequest, resp.Message)
		return
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

func (h *ConfigurationHandler) HandleGetConfiguration(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]
	if userID == "" {
		utils.WriteError(w, http.StatusBadRequest, "UserID is required")
		return
	}

	tokenUserID, ok := r.Context().Value("userID").(string)
	if !ok || tokenUserID != userID {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	resp, err := h.gateway.GetConfiguration(r.Context(), &pb.GetConfigurationRequest{
		UserId: userID,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.WriteJSON(w, http.StatusOK, resp)
}

func (h *ConfigurationHandler) HandleDeleteConfiguration(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]
	if userID == "" {
		utils.WriteError(w, http.StatusBadRequest, "UserID is required")
		return
	}

	tokenUserID, ok := r.Context().Value("userID").(string)
	if !ok || tokenUserID != userID {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	resp, err := h.gateway.DeleteConfiguration(r.Context(), &pb.DeleteConfigurationRequest{
		UserId: userID,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if !resp.Success {
		utils.WriteError(w, http.StatusBadRequest, resp.Message)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]bool{"success": resp.Success})
}
