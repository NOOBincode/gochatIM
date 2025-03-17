package websocket

import (
	"GochatIM/internal/infrastructure/auth"
	"GochatIM/pkg/logger"
	"net/http"
)

type Handler struct {
	gateway 	*Gateway
	authService *auth.TokenService
}

func NewHandler(gateway *Gateway, authService *auth.TokenService) *Handler {
	return &Handler{
		gateway: gateway,
		authService: authService,
	}
}

func (h *Handler) HandleWebsocket(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == ""{
		http.Error(w, "没有提供令牌", http.StatusUnauthorized)
		return
	}
	claims,err := h.authService.ValidateToken(r.Context(), token)
	if err != nil {
		http.Error(w, "令牌无效", http.StatusUnauthorized)
		logger.Errorf("无效的认证令牌:%v",err)
		return
	}
	UserID := claims.UserID
	deviceID := r.URL.Query().Get("device_id")
	if deviceID == "" {
		deviceID = "web"
	}
	h.gateway.HandleWebsocket(w,r,UserID,deviceID)
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux){
	mux.HandleFunc("/ws", h.HandleWebsocket)
}



