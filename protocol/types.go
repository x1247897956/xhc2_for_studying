package protocol

const (
	MessageTypeRegister = "register"
	MessageTypeCheckin  = "checkin"
	MessageTypeResult   = "result"
)

const (
	TaskTypeNoop   = "noop"
	TaskTypeWhoami = "whoami"
	TaskTypeShell  = "shell"
)

const (
	TaskStatusPending = "pending"
	TaskStatusDone    = "done"
	TaskStatusFailed  = "failed"
)

const (
	ModeBeacon  = "beacon"
	ModeSession = "session"
)
