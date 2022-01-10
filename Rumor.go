package main

import (
	"VAA_Uebung1/pkg/Cluster"
	"fmt"
	"log"

	"github.com/hashicorp/memberlist"
)

type Rumors struct {
	RummorsMsg        Cluster.Message `json:rummormsg`
	RecievedRumorsNum int
	Believable        bool
}

type RumorsList struct {
	Node   memberlist.Node
	Rumors map[string]*Rumors
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

func (rl *RumorsList) GetRomors(romors string) *Rumors {
	return rl.Rumors[romors]
}

func (rl *RumorsList) GetRomorsMsg(rumors string) *Cluster.Message {
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

func (r *Rumors) AddMsg(msg Cluster.Message) {
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
