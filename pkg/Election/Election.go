package Election

import (
	"github.com/hashicorp/memberlist"
)

type RingMessage struct {
	I_Wait     bool
	RingSender map[string]*memberlist.Node `json:"ring_sender"`
}

func (r *RingMessage) AddSender(node memberlist.Node) {
	if _, ok := r.RingSender[node.Name]; ok {
		return
	}
	r.RingSender[node.Name] = &node
}

func (r *RingMessage) ContainsNode(node memberlist.Node) bool {
	if _, ok := r.RingSender[node.Name]; ok {
		return true
	}
	return false
}

type ElectionExplorer struct {
	M            int                         `json:"m"`
	Initiator    *memberlist.Node            `json:"initiator"`
	RecievedFrom map[string]*memberlist.Node `json:"recievedfrom"`
	OtherSender  map[string]*memberlist.Node `json:"othersender"`
}

func (e *ElectionExplorer) Add_ID(id int) {
	e.M = id
}

func (e *ElectionExplorer) Add_OtherSender(node memberlist.Node) {
	if _, ok := e.OtherSender[node.Name]; ok {
		return
	}
	e.OtherSender[node.Name] = &node
}

func (e *ElectionExplorer) Add_RecievedFrom(node memberlist.Node) {
	if _, ok := e.RecievedFrom[node.Name]; ok {
		return
	}
	e.RecievedFrom[node.Name] = &node
}

func (e *ElectionExplorer) Clear() {
	e.M = -1
	e.Initiator = nil
	e.RecievedFrom = make(map[string]*memberlist.Node)
	e.OtherSender = make(map[string]*memberlist.Node)
}

func New() *ElectionExplorer {
	return &ElectionExplorer{
		OtherSender:  make(map[string]*memberlist.Node),
		RecievedFrom: make(map[string]*memberlist.Node),
	}
}

func NewElection(id int, node memberlist.Node) *ElectionExplorer {
	return &ElectionExplorer{
		M:            id,
		Initiator:    &node,
		OtherSender:  make(map[string]*memberlist.Node),
		RecievedFrom: make(map[string]*memberlist.Node),
	}
}

func (e *ElectionExplorer) ContainsNodeInRecievedFrom(node *memberlist.Node) bool {
	for _, sender := range e.RecievedFrom {
		if node.Name == sender.Name {
			return true
		}
	}
	return false
}
