package process_queue

type ProcessQueue interface {
	Process() error
}
