package Cluster

import (
	"fmt"
	"log"

	"github.com/hashicorp/memberlist"
)

type Rumors struct {
	RummorsMsg        Message `json:"rumorsmsg"`
	RecievedFrom      []memberlist.Node
	RecievedRumorsNum int  `json:"recievednum"`
	Believable        bool `json:"blievalbe"`
}

type RumorsList struct {
	Node   memberlist.Node    `json:"node"`
	Rumors map[string]*Rumors `json:"rumors"`
}

func (rl *RumorsList) AddRumorsToList(rumors Rumors) {
	if _, ok := rl.Rumors[rumors.RummorsMsg.Msg]; ok {
		return
	}
	temp_rumors := NewRumors(rumors)
	rl.Rumors[rumors.RummorsMsg.Msg] = temp_rumors
}

//if rumors is already recieved then the rumors recievedNum it will return true
func (rl *RumorsList) ContainsRumors(rumors *Rumors) bool {
	if rumors == nil {
		log.Println("The rumor message is empty!")
		return false
	}

	if _, ok := rl.Rumors[rumors.RummorsMsg.Msg]; ok {
		return true
	}

	return false
}

func (rl *RumorsList) IfRumorsIncrementRN(rumors *Rumors, blievalbeNum int) bool {
	if rl.ContainsRumors(rumors) {
		rl.Rumors[rumors.RummorsMsg.Msg].IncrementRecievedNum()
		if rl.Rumors[rumors.RummorsMsg.Msg].RecievedRumorsNum >= blievalbeNum {
			rl.Rumors[rumors.RummorsMsg.Msg].Believable = true
		}
		return true
	}
	return false
}

func (rl *RumorsList) AddRecievedFrom(rumors *Rumors, node memberlist.Node) {
	if node.Name == "" {
		log.Println("Could not add recieved from: node ist empty")
		return
	}
	rl.Rumors[rumors.RummorsMsg.Msg].RecievedFrom =
		append(rl.Rumors[rumors.RummorsMsg.Msg].RecievedFrom, node)
}

func (rl *RumorsList) GetRomors(romors string) *Rumors {
	return rl.Rumors[romors]
}

func (rl *RumorsList) GetRomorsMsg(rumors string) *Message {
	return &rl.GetRomors(rumors).RummorsMsg
}

func (rl *RumorsList) String() string {
	out := ""
	out += fmt.Sprintf("\t%s\n", rl.Node.Name)
	for rumors := range rl.Rumors {
		out += fmt.Sprintf("\tRumors: %s\tRecieved num: %d\tBlievable: %v\n", rl.GetRomors(rumors).RummorsMsg.Msg,
			rl.GetRomors(rumors).RecievedRumorsNum,
			rl.GetRomors(rumors).Believable,
		)
	}
	return out
}

func NewRumors(rumors Rumors) *Rumors {
	return &Rumors{
		RummorsMsg: rumors.RummorsMsg,
	}
}

func (r *Rumors) AddMsg(msg Message) {
	r.RummorsMsg = msg
}

func (r *Rumors) IncrementRecievedNum() {
	r.RecievedRumorsNum += 1
}

func (r *Rumors) GetRecievedNum() int {
	return r.RecievedRumorsNum
}

func NewRumorsList() *RumorsList {
	return &RumorsList{
		Rumors: map[string]*Rumors{},
	}
}

// func main() {
// 	// rl := NewRumorsList()

// 	r1 := Rumors{RummorsMsg: Cluster.Message{Msg: "First Rumors"}}
// 	r2 := Rumors{RummorsMsg: Cluster.Message{Msg: "Second Rumors"}}
// 	r3 := Rumors{RummorsMsg: Cluster.Message{Msg: "Third Rumors"}}

// 	rl := NewRumorsList()
// 	rl.AddRumorsToList(r1)
// 	rl.AddRumorsToList(r2)
// 	rl.AddRumorsToList(r3)

// 	r1.RummorsMsg = Cluster.Message{Msg: "Test Rumor"}

// 	fmt.Println(rl.String())

// 	rl.IfRumorsIncrementRN(&r2, 2)
// 	rl.IfRumorsIncrementRN(&r2, 2)
// 	fmt.Println(rl.String())

// }
