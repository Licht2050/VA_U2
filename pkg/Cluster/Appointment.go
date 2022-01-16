package Cluster

import (
	"github.com/hashicorp/memberlist"
)

type Appointment struct {
	Time    int             `json:"appointment_time"`
	Inviter memberlist.Node `json:"appointment_inviter"`
	Message Message         `json:"appointment_message"`
}

func (a *Appointment) Create_Available_Time(time int, node memberlist.Node) {
	a.Time = time
	a.Inviter = node
}

func (a *Appointment) Clear() Appointment {
	return Appointment{}
}

func (a *Appointment) Make_an_Appointment(recieved_time int, node memberlist.Node) {
	if recieved_time <= 0 {
		return
	}
	a.Time = a.Time / recieved_time
	a.Inviter = node
}

type Appointment_Protokoll struct {
	Self                   memberlist.Node
	A_Max                  int
	Appointments           map[string]Appointment
	Available_Appointments []int
}
