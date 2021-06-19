package tommyqueue

// Task represents a unit of work to be performed.
type Task struct {
	// Type indicates the type of task to be performed.
	Type string

	// Payload holds data needed to perform the task.
	Payload Payload
}

// NewTask returns a new Task given a type name and payload data.
//
// The payload values must be serializable.
func NewTask(typename string, payload map[string]interface{}) *Task {
	return &Task{
		Type:    typename,
		Payload: Payload{payload},
	}
}

// Payload holds arbitrary data needed for task execution.
type Payload struct {
	data map[string]interface{}
}
