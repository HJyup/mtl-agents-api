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

type UserGateway interface {
	CreatUser(ctx context.Context, payload *pb.CreateUserRequest) (*pb.CreateUserResponse, error)
	AuthUser(ctx context.Context, payload *pb.AuthUserRequest) (*pb.AuthUserResponse, error)
	GetUser(context.Context, *pb.GetUserRequest) (*pb.GetUserResponse, error)
	DeleteUser(context.Context, *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error)
}

type UserHandler struct {
	gateway UserGateway
}

func NewUserHandler(gateway UserGateway) *UserHandler {
	return &UserHandler{gateway: gateway}
}

func (h *UserHandler) RegisterRoutes(router *mux.Router) {
	userRouter := router.PathPrefix("/api/v1/users").Subrouter()
	userRouter.HandleFunc("/sign-up", h.HandleCreateUser).Methods("POST")
	userRouter.Handle("/sign-in", utils.TokenAuthMiddleware(http.HandlerFunc(h.HandleAuthUser))).Methods("GET")
	userRouter.Handle("/{userId}", utils.TokenAuthMiddleware(http.HandlerFunc(h.HandleGetUser))).Methods("GET")
	userRouter.Handle("/{userId}", utils.TokenAuthMiddleware(http.HandlerFunc(h.HandleDeleteUser))).Methods("DELETE")
}

func (h *UserHandler) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	var reqBody models.CreateUserRequest

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

	resp, err := h.gateway.CreatUser(r.Context(), &pb.CreateUserRequest{
		Username: reqBody.UserName,
		Email:    reqBody.Email,
		Password: reqBody.Password,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.WriteJSON(w, http.StatusCreated, resp)
}

func (h *UserHandler) HandleAuthUser(w http.ResponseWriter, r *http.Request) {
	var reqBody models.AuthenticateUserRequest

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

	resp, err := h.gateway.AuthUser(r.Context(), &pb.AuthUserRequest{
		Email:    reqBody.Email,
		Password: reqBody.Password,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.WriteJSON(w, http.StatusCreated, resp)
}

func (h *UserHandler) HandleGetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId := vars["userId"]
	if userId == "" {
		utils.WriteError(w, http.StatusBadRequest, "UserID is required")
		return
	}

	resp, err := h.gateway.GetUser(r.Context(), &pb.GetUserRequest{
		UserId: userId,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.WriteJSON(w, http.StatusOK, resp)
}

func (h *UserHandler) HandleDeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId := vars["userId"]
	if userId == "" {
		utils.WriteError(w, http.StatusBadRequest, "UserID is required")
		return
	}

	tokenUserID, ok := r.Context().Value("userID").(string)
	if !ok || tokenUserID != userId {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	resp, err := h.gateway.DeleteUser(r.Context(), &pb.DeleteUserRequest{
		UserId: userId,
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
