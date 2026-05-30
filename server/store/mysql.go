package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"xhc2_for_studying/protocol"
	"xhc2_for_studying/server/core"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Stores groups the persistence backends shared by server components.
type Stores struct {
	Beacons  BeaconStore
	Tasks    ServerTaskStore
	Sessions *SessionStore
	Implants ImplantStore
}

// NewMemoryStores returns the default in-memory persistence set.
func NewMemoryStores() Stores {
	return Stores{
		Beacons:  NewBeaconStore(),
		Tasks:    NewServerTaskStore(),
		Sessions: NewSessionStore(),
		Implants: NewImplantStore(),
	}
}

// NewMySQLStores opens a MySQL database through GORM, migrates required tables,
// and returns stores backed by the database. Sessions remain in memory because
// they contain live cipher contexts.
func NewMySQLStores(dsn string) (Stores, *sql.DB, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return Stores{}, nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return Stores{}, nil, err
	}
	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		return Stores{}, nil, err
	}
	stores, err := NewGORMStores(db)
	if err != nil {
		sqlDB.Close()
		return Stores{}, nil, err
	}
	return stores, sqlDB, nil
}

// NewGORMStores migrates schema and returns stores backed by an existing GORM
// database handle. Tests use this with SQLite; production uses MySQL.
func NewGORMStores(db *gorm.DB) (Stores, error) {
	if err := db.AutoMigrate(&beaconModel{}, &taskModel{}, &implantModel{}); err != nil {
		return Stores{}, err
	}
	return Stores{
		Beacons:  NewGORMBeaconStore(db),
		Tasks:    NewGORMTaskStore(db),
		Sessions: NewSessionStore(),
		Implants: NewGORMImplantStore(db),
	}, nil
}

type beaconModel struct {
	ID              string `gorm:"primaryKey;size:128"`
	Hostname        string
	Username        string
	OS              string `gorm:"column:os;size:64"`
	Arch            string `gorm:"size:64"`
	IntervalSeconds int64
	JitterSeconds   int64
	LastCheckIn     int64
	RemoteAddress   string
}

func (beaconModel) TableName() string {
	return "beacons"
}

type taskModel struct {
	TaskID        string `gorm:"primaryKey;size:128"`
	Type          string `gorm:"size:64"`
	ImplantID     string `gorm:"index:idx_tasks_implant_status,priority:1;size:128"`
	Payload       string
	Status        string `gorm:"index:idx_tasks_implant_status,priority:2;size:64"`
	ResultJSON    string `gorm:"type:json"`
	Error         string
	CreatedUnix   int64
	CompletedUnix *int64
}

func (taskModel) TableName() string {
	return "tasks"
}

type implantModel struct {
	PubKeyDigest         string `gorm:"primaryKey;size:128"`
	ImplantAgePrivateKey string
	ExtMapJSON           string `gorm:"type:json"`
}

func (implantModel) TableName() string {
	return "implants"
}

type GORMBeaconStore struct {
	db *gorm.DB
}

func NewGORMBeaconStore(db *gorm.DB) *GORMBeaconStore {
	return &GORMBeaconStore{db: db}
}

func (s *GORMBeaconStore) Add(beacon *core.Beacon) error {
	if beacon == nil || beacon.ID == "" {
		return nil
	}
	model := beaconToModel(beacon)
	return s.db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&model).Error
}

func (s *GORMBeaconStore) Get(id string) (*core.Beacon, error) {
	var model beaconModel
	if err := s.db.First(&model, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrBeaconNotFound
		}
		return nil, err
	}
	return modelToBeacon(model), nil
}

func (s *GORMBeaconStore) UpdateCheckIn(id string, lastCheckIn int64, remoteAddress string) error {
	updates := map[string]any{"last_check_in": lastCheckIn}
	if remoteAddress != "" {
		updates["remote_address"] = remoteAddress
	}
	result := s.db.Model(&beaconModel{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrBeaconNotFound
	}
	return nil
}

func (s *GORMBeaconStore) ListIDs() []string {
	var ids []string
	if err := s.db.Model(&beaconModel{}).Pluck("id", &ids).Error; err != nil {
		return nil
	}
	return ids
}

type GORMTaskStore struct {
	db *gorm.DB
}

func NewGORMTaskStore(db *gorm.DB) *GORMTaskStore {
	return &GORMTaskStore{db: db}
}

func (s *GORMTaskStore) AddTask(task *core.ServerTask) error {
	if task == nil || task.TaskID == "" {
		return nil
	}
	model := taskToModel(task)
	return s.db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&model).Error
}

func (s *GORMTaskStore) GetTask(taskID string) (*core.ServerTask, error) {
	var model taskModel
	if err := s.db.First(&model, "task_id = ?", taskID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrServerTaskNotFound
		}
		return nil, err
	}
	return modelToTask(model), nil
}

func (s *GORMTaskStore) GetPendingTasksByImplantID(implantID string) []*core.ServerTask {
	var models []taskModel
	if err := s.db.Where("implant_id = ? AND status = ?", implantID, protocol.TaskStatusPending).Find(&models).Error; err != nil {
		return nil
	}
	tasks := make([]*core.ServerTask, 0, len(models))
	for _, model := range models {
		tasks = append(tasks, modelToTask(model))
	}
	return tasks
}

func (s *GORMTaskStore) UpdateTask(taskID string, taskRes protocol.TaskResult) error {
	completed := taskRes.Completed
	if completed.IsZero() {
		completed = time.Now()
		taskRes.Completed = completed
	}
	completedUnix := completed.Unix()
	result := s.db.Model(&taskModel{}).Where("task_id = ?", taskID).Updates(map[string]any{
		"status":         taskRes.Status,
		"result_json":    marshalTaskResult(taskRes),
		"error":          taskRes.Error,
		"completed_unix": &completedUnix,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrServerTaskNotFound
	}
	return nil
}

type GORMImplantStore struct {
	db *gorm.DB
}

func NewGORMImplantStore(db *gorm.DB) *GORMImplantStore {
	return &GORMImplantStore{db: db}
}

func (s *GORMImplantStore) Set(pubKeyDigest string, record *ImplantRecord) error {
	if pubKeyDigest == "" || record == nil {
		return nil
	}
	extMapJSON, err := json.Marshal(record.ExtMap)
	if err != nil {
		return err
	}
	model := implantModel{
		PubKeyDigest:         pubKeyDigest,
		ImplantAgePrivateKey: record.ImplantAgePrivateKey,
		ExtMapJSON:           string(extMapJSON),
	}
	return s.db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&model).Error
}

func (s *GORMImplantStore) Get(pubKeyDigest string) (*ImplantRecord, bool) {
	var model implantModel
	if err := s.db.First(&model, "pub_key_digest = ?", pubKeyDigest).Error; err != nil {
		return nil, false
	}
	record := &ImplantRecord{ImplantAgePrivateKey: model.ImplantAgePrivateKey}
	if err := json.Unmarshal([]byte(model.ExtMapJSON), &record.ExtMap); err != nil {
		return nil, false
	}
	return record, true
}

func beaconToModel(beacon *core.Beacon) beaconModel {
	return beaconModel{
		ID:              beacon.ID,
		Hostname:        beacon.Hostname,
		Username:        beacon.Username,
		OS:              beacon.OS,
		Arch:            beacon.Arch,
		IntervalSeconds: beacon.Interval,
		JitterSeconds:   beacon.Jitter,
		LastCheckIn:     beacon.LastCheckIn,
		RemoteAddress:   beacon.RemoteAddress,
	}
}

func modelToBeacon(model beaconModel) *core.Beacon {
	return &core.Beacon{
		ID:            model.ID,
		Hostname:      model.Hostname,
		Username:      model.Username,
		OS:            model.OS,
		Arch:          model.Arch,
		Interval:      model.IntervalSeconds,
		Jitter:        model.JitterSeconds,
		LastCheckIn:   model.LastCheckIn,
		RemoteAddress: model.RemoteAddress,
	}
}

func taskToModel(task *core.ServerTask) taskModel {
	created := task.CreatedAt
	if created.IsZero() {
		created = time.Now()
	}
	return taskModel{
		TaskID:        task.TaskID,
		Type:          task.Type,
		ImplantID:     task.ImplantID,
		Payload:       task.Payload,
		Status:        task.Status,
		ResultJSON:    marshalOptionalTaskResult(task.Result),
		Error:         task.Error,
		CreatedUnix:   created.Unix(),
		CompletedUnix: unixPtrOrNil(task.CompletedAt),
	}
}

func modelToTask(model taskModel) *core.ServerTask {
	task := &core.ServerTask{
		TaskID:    model.TaskID,
		Type:      model.Type,
		ImplantID: model.ImplantID,
		Payload:   model.Payload,
		Status:    model.Status,
		Error:     model.Error,
		CreatedAt: time.Unix(model.CreatedUnix, 0),
	}
	if model.CompletedUnix != nil {
		task.CompletedAt = time.Unix(*model.CompletedUnix, 0)
	}
	if model.ResultJSON != "" {
		_ = json.Unmarshal([]byte(model.ResultJSON), &task.Result)
	}
	return task
}

func marshalTaskResult(taskRes protocol.TaskResult) string {
	data, _ := json.Marshal(taskRes)
	return string(data)
}

func marshalOptionalTaskResult(taskRes protocol.TaskResult) string {
	if taskRes.TaskID == "" && taskRes.ImplantID == "" && taskRes.Status == "" && taskRes.Error == "" && taskRes.Output == "" && taskRes.Completed.IsZero() {
		return ""
	}
	return marshalTaskResult(taskRes)
}

func unixPtrOrNil(t time.Time) *int64 {
	if t.IsZero() {
		return nil
	}
	value := t.Unix()
	return &value
}
