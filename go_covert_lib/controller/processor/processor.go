package processor

type Processor interface {
	Process(data []byte) ([]byte, error)
	Unprocess(data []byte) ([]byte, error)
}
