package queue

type Task struct {
	ID      string
	Lane    string
	Payload string
}

type Lane struct {
	Name        string
	TaskHandler TaskHandler
}

// TaskHandler processes one task's kind/payload. Registered per lane.
type TaskHandler func(payload string) error
