package protocol

// Well-known message type constants used to classify messages exchanged
// between implant and server.
const (
	MessageTypeRegister = "register"
	MessageTypePoll     = "poll"
	MessageTypeResult   = "result"
)

// Task type constants understood by the implant.
const (
	TaskTypeNoop   = "noop"
	TaskTypeWhoami = "whoami"
	TaskTypeShell  = "shell"
)

// Task status constants representing the lifecycle of a task.
const (
	TaskStatusPending = "pending"
	TaskStatusDone    = "done"
	TaskStatusFailed  = "failed"
)

// Operating modes for the implant communication loop.
const (
	ModeBeacon  = "beacon"
	ModeSession = "session"
)
