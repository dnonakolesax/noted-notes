package services

type SocketService struct {

}

func (ss *SocketService) Compile() error {
	return nil
}

func (ss *SocketService) Run() (string, error) {
	return "", nil

}

func (ss *SocketService) Update() error {
	return nil
}

func NewSocketService() *SocketService {
	return &SocketService{}
}