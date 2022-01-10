package Cluster

import (
	"VAA_Uebung1/pkg/Graph"
	"VAA_Uebung1/pkg/Neighbour"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/memberlist"
)

type SyncerDelegate struct {
	MasterNode     *memberlist.Node
	LocalNode      *memberlist.Node
	Node           *memberlist.Memberlist
	Neighbours     *Neighbour.NeighboursList
	NodesNeighbour *Neighbour.NodesAndNeighbours
	NeighbourNum   *int
	NodeList       *Neighbour.NodesList
	Graph          *Graph.Graph
	RumorsList     *RumorsList
	// Broadcasts        *memberlist.TransmitLimitedQueue
	neighbourFilePath    *string
	BelievableRumorsRNum *int
}

//compare the incoming byte message to structs
func CompareJson(msg []byte, NeigbourGraph interface{}) string {
	var receivedMsg map[string]interface{}
	err := json.Unmarshal(msg, &receivedMsg)
	_ = err

	// emptyValue := reflect.ValueOf(NeigbourGraph).Type()
	for key := range receivedMsg {

		if key == "neighbours" {
			return "neighbourStrucht"

		}
		if key == "rumorsmsg" {
			return "rumors"
		}
	}
	return "msg"
}

// NotifyMsg is called when a user-data message is received.
// Care should be taken that this method does not block, since doing
// so would block the entire UDP packet receive loop. Additionally, the byte
// slice may be modified after the call returns, so it should be copied if needed.
func (sd *SyncerDelegate) NotifyMsg(msg []byte) {
	// fmt.Println(string(msg))
	// validate if the msg is struct of Neigbour{}

	check := CompareJson(msg, Neighbour.NeighboursList{})

	if check == "msg" {

		// ms := NeigbourGraph{Node: sd.Node.LocalNode().Name}
		var receivedMsg Message
		err := json.Unmarshal(msg, &receivedMsg)
		_ = err

		if receivedMsg.Msg == "leave" {
			//Leave will kill the process
			//and the node will remove from Cluster-Memberlist
			sd.Leave()

		} else if receivedMsg.Msg == "readNeighbour" {
			fmt.Println("Got message -----------------------------------------------")
			//clear the available neighbour list
			//Read Graph from file
			//add Nodes to Neighbourlist if there is a releastionship for this nod found
			//Send the new neighbour list to MasterNode
			sd.neighbourFilePath = &receivedMsg.FilePath
			ReadNeighbourFromDot(sd)

			// sd.Broadcast()
			// sd.Node.UpdateNode(time.Millisecond)

		}

	} else if check == "rumors" {
		fmt.Println("GEREUCHTTTTTTTTTTTTTTTTTTTTTTTTTTTT")
		rumors := new(Rumors)
		err := json.Unmarshal(msg, rumors)
		_ = err

		if !sd.RumorsList.IfRumorsIncrementRN(rumors, *sd.BelievableRumorsRNum) {
			sd.RumorsList.AddRumorsToList(*rumors)
			sd.RumorsList.AddRecievedFrom(rumors, rumors.RecievedFrom[0])
			rumors.RecievedFrom[0] = *sd.LocalNode
			sd.SendMsgToNeighbours(rumors)
		} else {
			sd.RumorsList.AddRecievedFrom(rumors, rumors.RecievedFrom[0])
		}

		fmt.Printf("Rumors: %s\n\tcount: %d\tBleivalbe: %v\t\n\tRecieved from: %v\n", sd.RumorsList.Rumors[rumors.RummorsMsg.Msg].RummorsMsg.Msg,
			sd.RumorsList.Rumors[rumors.RummorsMsg.Msg].RecievedRumorsNum,
			sd.RumorsList.Rumors[rumors.RummorsMsg.Msg].Believable,
			sd.RumorsList.Rumors[rumors.RummorsMsg.Msg].RecievedFrom,
		)

	} else if check == "neighbourStrucht" {
		//MasterNode recieve's neighbours from every node in the cluster afther any update occurred
		//afther recieved the message it will insert nodes and their neighbour's in to "NodesAndNeighbours" list
		if sd.Node.LocalNode().Name == "Master" {

			var receivedMsg Neighbour.NeighboursList
			err := json.Unmarshal(msg, &receivedMsg)
			errorAnd_Msg := Error_And_Msg{Err: err, Text: "Could not encode the NeighboursInfo message"}
			Check(errorAnd_Msg)

			sd.NodesNeighbour.AddNodesAndNeighbours(receivedMsg)
		}
	}
}

func (sd *SyncerDelegate) SendMsgToNeighbours(value interface{}) {
	if rumors, ok := value.(*Rumors); ok {
		recievedFrom_temp := rumors.RummorsMsg.Snder
		rumors.RummorsMsg.Snder = sd.LocalNode.Name
		rumors.RecievedFrom = append(rumors.RecievedFrom, *sd.LocalNode)

		for _, neighbour := range sd.Neighbours.Neighbours {

			if neighbour.Name != recievedFrom_temp {
				rumors.RummorsMsg.Receiver = neighbour.Name
				rumors.RummorsMsg.SendTime = time.Now()
				body, err := json.Marshal(rumors)
				error_and_msg := Error_And_Msg{Err: err, Text: "Encode the rumors faild!"}
				Check(error_and_msg)

				sd.Node.SendBestEffort(&neighbour, body)
			}

		}
	}
}

func ReadNeighbourFromDot(sd *SyncerDelegate) {
	if sd.neighbourFilePath != nil && *sd.neighbourFilePath != "" {

		g := Graph.NewDiGraph()
		g.ParseFileToGraph(*sd.neighbourFilePath)

		if AddNodesToNeighbourList(g, sd) {
			body, _ := json.Marshal(sd.Neighbours)
			sd.Node.SendBestEffort(sd.MasterNode, body)
		}
	}
}

func AddNodesToNeighbourList(g *Graph.Graph, sd *SyncerDelegate) bool {
	for _, node := range g.Nodes {
		if node.Name == sd.LocalNode.Name {

			neighbours := g.GetEdges(node.Name)
			if len(neighbours.Nodes) > 0 {
				sd.Neighbours.ClearNeighbours()
				for _, neighbour := range neighbours.Nodes {
					//it add to the neighbour list if the node is a cluster memeber
					found_Node := SearchMemberbyName(neighbour.Name, sd.Node)
					if found_Node.Name == neighbour.Name {

						sd.Neighbours.AddNeighbour(*found_Node)
					}
				}
				return true
			}
			return false
		}
	}
	return false
}

func (d *SyncerDelegate) NotifyJoin(node *memberlist.Node) {

	d.NodeList.AddNode(node)
	d.Neighbours.UpdateNeighbourList(*d.NeighbourNum, *d.NodeList)
	body, _ := json.Marshal(d.Neighbours)
	d.Node.SendBestEffort(d.MasterNode, body)

	log.Printf("notify join %s(%s:%d)", node.Name, node.Addr.To4().String(), node.Port)

	fmt.Printf("%s Neigbour's: \n", d.Neighbours.Node.Name)
	for _, p := range d.Neighbours.Neighbours {
		fmt.Println("Neigbour---------: ", p.Name)
	}

}

func (d *SyncerDelegate) NotifyLeave(node *memberlist.Node) {

	log.Printf("notify leave %s(%s:%d)", node.Name, node.Addr.To4().String(), node.Port)

	if d.LocalNode.Name == "Master" {
		d.NodesNeighbour.RemoveNodesNeighbours(*node)

	}

	d.NodeList.RemoveNode(node)
	if d.Neighbours.Contains(node) {
		d.Neighbours.UpdateNeighbourList(*d.NeighbourNum, *d.NodeList)
		body, _ := json.Marshal(d.Neighbours)
		d.Node.SendBestEffort(d.MasterNode, body)
	}

}
func (d *SyncerDelegate) NotifyUpdate(node *memberlist.Node) {

	log.Printf("notify update %s(%s:%d)", node.Name, node.Addr.To4().String(), node.Port)
}

func BroadcastClusterMessage(ml *memberlist.Memberlist, msg *Message) {
	if msg == nil {
		errMessage := "Could not broadcast an empty message"
		log.Println(errMessage)
	}

	body, err := json.Marshal(msg)
	error_and_Message := Error_And_Msg{Err: err, Text: "Could not encode and broadcast the message"}
	Check(error_and_Message)

	for _, mem := range ml.Members() {
		// if mem.Name == "Node02" {
		ml.SendBestEffort(mem, body)
		// }
	}
}

// GetBroadcasts is called when user data messages can be broadcast.
// It can return a list of buffers to send. Each buffer should assume an
// overhead as provided with a limit on the total byte size allowed.
// The total byte size of the resulting data to send must not exceed
// the limit.
func (sd *SyncerDelegate) GetBroadcasts(overhead, limit int) [][]byte {
	return [][]byte{}
}

// func (sd *SyncerDelegate) QueueBroadcast(msg []byte) {
// 	sd.Broadcasts.QueueBroadcast(&MemberlistBroadcast{"test", msg})
// }

// LocalState is used for a TCP Push/Pull. This is sent to
// the remote side in addition to the membership information. Any
// data can be sent here. See MergeRemoteState as well. The `join`
// boolean indicates this is for a join instead of a push/pull.
func (sd *SyncerDelegate) LocalState(join bool) []byte {
	return nil
}

// MergeRemoteState is invoked after a TCP Push/Pull. This is the
// state received from the remote side and is the result of the
// remote side's LocalState call. The 'join'
// boolean indicates this is for a join instead of a push/pull.
func (sd *SyncerDelegate) MergeRemoteState(buf []byte, join bool) {

}

// NodeMeta is used to retrieve meta-data about the current node
// when broadcasting an alive message. It's length is limited to
// the given byte size. This metadata is available in the Node structure.
func (sd *SyncerDelegate) NodeMeta(limit int) []byte {
	return []byte{}
}
