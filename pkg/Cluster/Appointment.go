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
	a.Time = (a.Time + recieved_time) / 2
	a.Inviter = node
}

type Appointment_Protocol struct {
	Self                     memberlist.Node
	A_Max                    int
	Counter                  int
	Appointments             map[string]Appointment
	Available_Appointments   []int
	Rand_Selected_Neighbours map[string]memberlist.Node
}

func CreateAppointmentProtocol(self memberlist.Node, aMax int, avialable_ap []int) *Appointment_Protocol {
	return &Appointment_Protocol{
		Self:                     self,
		A_Max:                    aMax,
		Counter:                  0,
		Appointments:             map[string]Appointment{},
		Available_Appointments:   avialable_ap,
		Rand_Selected_Neighbours: make(map[string]memberlist.Node),
	}
}

func (ap *Appointment_Protocol) CopyRandSelected_Neighbour(neighbour_map map[string]memberlist.Node) {
	for _, neighbour := range neighbour_map {
		if _, ok := ap.Rand_Selected_Neighbours[neighbour.Name]; !ok {
			ap.Rand_Selected_Neighbours[neighbour.Name] = neighbour
		}
	}
}

func (ap *Appointment_Protocol) Add_Appointment(a Appointment) {
	if _, ok := ap.Appointments[a.Inviter.Name]; !ok {
		ap.Appointments[a.Inviter.Name] = a
	}
}

func (ap *Appointment_Protocol) Start_Value() {
	available_Appointment := []int{1, 2, 3, 4, 7, 8, 9, 10, 11, 12}
	ap.A_Max = 3
	ap.Counter = 0
	ap.Available_Appointments = available_Appointment
	ap.Appointments = map[string]Appointment{}
}
