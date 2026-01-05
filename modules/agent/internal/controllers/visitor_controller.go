package controllers

import (
	"encoding/json"
	"net/http"
	"strings"

	"agent/internal/services"
)

type VisitorController struct {
	visitorService *services.VisitorService
}

func NewVisitorController(vs *services.VisitorService) *VisitorController {
	return &VisitorController{visitorService: vs}
}

func (c *VisitorController) Notify(w http.ResponseWriter, r *http.Request) {
	var info services.VisitorInfo
	json.NewDecoder(r.Body).Decode(&info)

	info.IP = getClientIP(r)
	info.UserAgent = r.Header.Get("User-Agent")
	info.Referrer = r.Header.Get("Referer")

	go c.visitorService.NotifyVisitor(info)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func getClientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return strings.TrimSpace(strings.Split(ip, ",")[0])
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	host := r.RemoteAddr
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		return host[:idx]
	}
	return host
}
