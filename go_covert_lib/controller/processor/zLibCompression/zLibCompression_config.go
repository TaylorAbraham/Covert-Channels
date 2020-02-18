package zLibCompression

type ConfigClient struct {
}

func GetDefault() ConfigClient {
	return ConfigClient{}
}

func ToProcessor(cc ConfigClient) (*ZLibCompression, error) {
	return &ZLibCompression{}, nil
}
