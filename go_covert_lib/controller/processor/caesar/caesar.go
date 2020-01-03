package caesar

type Caesar struct {
	Shift int8
}

func (c *Caesar) Process(data []byte) ([]byte, error) {
	var newData []byte = make([]byte, len(data))
	for i := range newData {
			newData[i] = data[i] + byte(c.Shift)
	}
	return newData[:], nil
}

func (c *Caesar) Unprocess(data []byte) ([]byte, error) {
	var newData []byte = make([]byte, len(data))
	for i := range newData {
			newData[i] = data[i] - byte(c.Shift)
	}
	return newData[:], nil
}
