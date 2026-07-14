package system_service

import fifo "github.com/foize/go.fifo"

var GlobalSystemService *SystemService

type SystemService struct {
	EventQueue *fifo.Queue
}

func NewSystemService() *SystemService {
	return &SystemService{
		EventQueue: fifo.NewQueue(),
	}
}

func SetupSystemService() {
	GlobalSystemService = NewSystemService()
}
