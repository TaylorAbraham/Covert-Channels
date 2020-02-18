package gZipCompression

type ConfigClient struct {
}

func GetDefault() ConfigClient {
	return ConfigClient{}
}

func ToProcessor(cc ConfigClient) (*GZipCompression, error) {
	return &GZipCompression{}, nil
}
