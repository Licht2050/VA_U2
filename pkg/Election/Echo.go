package Election

import "github.com/hashicorp/memberlist"

type Echo struct {
	Coordinator     int                         `json:"coordinator"`
	EchoSender      memberlist.Node             `json:"echo_sender"`
	EchoSenderList  map[string]*memberlist.Node `json:"echo_sender_List"`
	EchoRecievedNum int
	EchoWaitedNum   int
}

func NewEcho(coordinator int, node memberlist.Node) *Echo {
	return &Echo{
		Coordinator:    coordinator,
		EchoSender:     node,
		EchoSenderList: make(map[string]*memberlist.Node),
	}
}

func (e *Echo) AddSender(node memberlist.Node) {
	if _, ok := e.EchoSenderList[node.Name]; ok {
		return
	}
	e.EchoSenderList[node.Name] = &node
}

func (e *Echo) Clear() {
	e.Coordinator = 0
	e.EchoRecievedNum = 0
	e.EchoWaitedNum = 0
	e.EchoSenderList = make(map[string]*memberlist.Node)
}
