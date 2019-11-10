package channel

type Channel interface {
	Receive(data []byte, progress chan<- uint64) (uint64, error)
	Send(data []byte, progress chan<- uint64) (uint64, error)
	Close() error
}
