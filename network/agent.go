package network

type Agent interface {
	Run() error
	OnClose() error
}
