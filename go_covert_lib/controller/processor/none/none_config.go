package none

type ConfigClient struct {
}

func GetDefault() ConfigClient {
	return ConfigClient{}
}

func ToProcessor (cc ConfigClient) (*None, error) {
	return &None{}, nil
}
