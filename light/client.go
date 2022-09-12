package light

type TwampLightClient struct{}

func NewLightClient() *TwampLightClient {
	return &TwampLightClient{}
}

func (c *TwampLightClient) Connect(hostname string, port int) (*TwampLightConnection, error) {
	twampConnection := NewTwampLightConnection(hostname, port)
	return twampConnection, nil
}
