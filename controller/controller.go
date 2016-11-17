package controller

type Controller interface {
	Start()
}

type controller struct {
	listenAddress string
}

func New(listenAddress string) Controller {
	return &controller{
		listenAddress: listenAddress,
	}
}

// Start starts the server.
func (c *controller) Start() {
	c.startServer()
}
