package light

type TwampLightClient struct{}

func NewLightClient() *TwampLightClient {
	return &TwampLightClient{}
}

func (c *TwampLightClient) Connect(hostname string, port uint16) (*TwampLightConnection, error) {
	twampConnection := NewTwampLightConnection(hostname, port)
	return twampConnection, nil
}
