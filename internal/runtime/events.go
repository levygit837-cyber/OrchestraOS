package runtime

const (
	EventCallStarted   = "runtime.call_started"
	EventCallCompleted = "runtime.call_completed"
	EventCallFailed    = "runtime.call_failed"

	EventStreamStarted   = "runtime.stream_started"
	EventStreamChunkRecv = "runtime.stream_chunk"
	EventStreamCompleted = "runtime.stream_completed"
	EventStreamFailed    = "runtime.stream_failed"
	EventRetryAttempted  = "runtime.retry_attempted"
)
