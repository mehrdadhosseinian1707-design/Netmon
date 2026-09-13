package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/netmon/netmon/internal/tools"
)

// ToolsHandler handles tool-related API requests
type ToolsHandler struct {
	registry *tools.ToolRegistry
}

func NewToolsHandler(registry *tools.ToolRegistry) *ToolsHandler {
	return &ToolsHandler{
		registry: registry,
	}
}

// RegisterRoutes registers tool API routes
func (h *ToolsHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/api/tools", h.ListTools).Methods("GET")
	r.HandleFunc("/api/tools/installed", h.ListInstalledTools).Methods("GET")
	r.HandleFunc("/api/tools/dependencies", h.CheckDependencies).Methods("GET")
	r.HandleFunc("/api/tools/{name}", h.GetToolInfo).Methods("GET")
	r.HandleFunc("/api/tools/{name}/execute", h.ExecuteTool).Methods("POST")
	r.HandleFunc("/api/tools/categories/{category}", h.GetToolsByCategory).Methods("GET")
}

// ListTools returns all available tools
func (h *ToolsHandler) ListTools(w http.ResponseWriter, r *http.Request) {
	tools := h.registry.GetAllToolsInfo()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tools": tools,
		"count": len(tools),
	})
}

// ListInstalledTools returns all installed tools
func (h *ToolsHandler) ListInstalledTools(w http.ResponseWriter, r *http.Request) {
	allTools := h.registry.List()
	installed := make([]tools.ToolInfo, 0)

	for _, tool := range allTools {
		if tool.IsInstalled() {
			if info, err := h.registry.GetToolInfo(tool.Name()); err == nil {
				installed = append(installed, *info)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tools": installed,
		"count": len(installed),
	})
}

// CheckDependencies checks all tool dependencies
func (h *ToolsHandler) CheckDependencies(w http.ResponseWriter, r *http.Request) {
	deps := h.registry.CheckDependencies()

	installed := 0
	missing := 0
	for _, isInstalled := range deps {
		if isInstalled {
			installed++
		} else {
			missing++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"dependencies": deps,
		"installed":    installed,
		"missing":      missing,
		"total":        len(deps),
	})
}

// GetToolInfo returns information about a specific tool
func (h *ToolsHandler) GetToolInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	info, err := h.registry.GetToolInfo(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

// ExecuteToolRequest represents a tool execution request
type ExecuteToolRequest struct {
	Args    []string               `json:"args"`
	Timeout int                    `json:"timeout,omitempty"` // seconds
	Options map[string]interface{} `json:"options,omitempty"`
}

// ExecuteTool executes a network analysis tool
func (h *ToolsHandler) ExecuteTool(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	var req ExecuteToolRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tool, err := h.registry.Get(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if !tool.IsInstalled() {
		http.Error(w, "Tool not installed: "+tool.InstallCommand(), http.StatusServiceUnavailable)
		return
	}

	// Set timeout (default 30 seconds)
	timeout := 30 * time.Second
	if req.Timeout > 0 {
		timeout = time.Duration(req.Timeout) * time.Second
	}

	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	// Execute tool
	result, err := tool.Execute(ctx, req.Args)
	if err != nil && result == nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Try to parse output
	if result.Success {
		parsed, err := tool.ParseOutput(result.Output)
		if err == nil {
			result.ParsedData = parsed
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetToolsByCategory returns all tools in a category
func (h *ToolsHandler) GetToolsByCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	category := tools.ToolCategory(vars["category"])

	categoryTools := h.registry.ListByCategory(category)
	infos := make([]tools.ToolInfo, 0)

	for _, tool := range categoryTools {
		if info, err := h.registry.GetToolInfo(tool.Name()); err == nil {
			infos = append(infos, *info)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"category": category,
		"tools":    infos,
		"count":    len(infos),
	})
}

// Specialized tool endpoints

// PingRequest represents a ping request
type PingRequest struct {
	Host  string `json:"host"`
	Count int    `json:"count"`
}

// Ping executes a ping test
func (h *ToolsHandler) Ping(w http.ResponseWriter, r *http.Request) {
	var req PingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	tool, err := h.registry.Get("ping")
	if err != nil {
		http.Error(w, "Ping tool not available", http.StatusServiceUnavailable)
		return
	}

	pingTool, ok := tool.(*tools.PingTool)
	if !ok {
		http.Error(w, "Invalid tool type", http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	result, err := pingTool.Ping(ctx, req.Host, req.Count)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// TracerouteRequest represents a traceroute request
type TracerouteRequest struct {
	Host    string `json:"host"`
	MaxHops int    `json:"max_hops,omitempty"`
}

// Traceroute executes a traceroute test
func (h *ToolsHandler) Traceroute(w http.ResponseWriter, r *http.Request) {
	var req TracerouteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.MaxHops == 0 {
		req.MaxHops = 30
	}

	tool, err := h.registry.Get("traceroute")
	if err != nil {
		http.Error(w, "Traceroute tool not available", http.StatusServiceUnavailable)
		return
	}

	tracerouteTool, ok := tool.(*tools.TracerouteTool)
	if !ok {
		http.Error(w, "Invalid tool type", http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	result, err := tracerouteTool.Traceroute(ctx, req.Host, req.MaxHops)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// DNSRequest represents a DNS lookup request
type DNSRequest struct {
	Domain     string `json:"domain"`
	RecordType string `json:"record_type,omitempty"`
	Server     string `json:"server,omitempty"`
}

// DNSLookup executes a DNS lookup
func (h *ToolsHandler) DNSLookup(w http.ResponseWriter, r *http.Request) {
	var req DNSRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	tool, err := h.registry.Get("dig")
	if err != nil {
		http.Error(w, "DNS tool not available", http.StatusServiceUnavailable)
		return
	}

	digTool, ok := tool.(*tools.DigTool)
	if !ok {
		http.Error(w, "Invalid tool type", http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	result, err := digTool.Dig(ctx, req.Domain, req.RecordType, req.Server)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// RegisterSpecializedRoutes registers specialized tool routes
func (h *ToolsHandler) RegisterSpecializedRoutes(r *mux.Router) {
	r.HandleFunc("/api/tools/ping", h.Ping).Methods("POST")
	r.HandleFunc("/api/tools/traceroute", h.Traceroute).Methods("POST")
	r.HandleFunc("/api/tools/dns", h.DNSLookup).Methods("POST")
}
