package none

type None struct{}

func (n *None) Process(data []byte) ([]byte, error) {
	return data, nil
}

func (n *None) Unprocess(data []byte) ([]byte, error) {
	return data, nil
}
