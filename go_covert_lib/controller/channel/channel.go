package channel

type Channel interface {
	Receive(data []byte) (uint64, error)
	Send(data []byte) (uint64, error)
	Close() error
}
