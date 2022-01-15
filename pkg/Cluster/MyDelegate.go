package Cluster

import (
	"VAA_Uebung1/pkg/Election"
	"VAA_Uebung1/pkg/Graph"
	"VAA_Uebung1/pkg/Neighbour"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/memberlist"
)

type SyncerDelegate struct {
	MasterNode       *memberlist.Node
	LocalNode        *memberlist.Node
	Node             *memberlist.Memberlist
	Neighbours       *Neighbour.NeighboursList
	NodesNeighbour   *Neighbour.NodesAndNeighbours
	NeighbourNum     *int
	NodeList         *Neighbour.NodesList
	Graph            *Graph.Graph
	RumorsList       *RumorsList
	ElectionExplorer *Election.ElectionExplorer
	EchoMessage      *Election.Echo
	RingMessage      *Election.RingMessage
	EchoCounter      *int
	//ElectionProtokol []*Election.ElectionExplorer
	// Broadcasts        *memberlist.TransmitLimitedQueue
	neighbourFilePath    *string
	BelievableRumorsRNum *int
}

//compare the incoming byte message to structs
func CompareJson(msg []byte, NeigbourGraph interface{}) int {
	var receivedMsg map[string]interface{}
	err := json.Unmarshal(msg, &receivedMsg)
	_ = err

	// emptyValue := reflect.ValueOf(NeigbourGraph).Type()
	for key := range receivedMsg {

		if key == "neighbours" {
			return NEIGHBOUR_INFO_MESSAGE

		}
		if key == "m" {
			return ELECTION_EXPLORER_MESSAGE
		}
		if key == "coordinator" {
			return ELECTION_ECHO_MESSAGE
		}
		if key == "ring_sender" {
			return -1
		}
	}
	return MESSAGE
}

// NotifyMsg is called when a user-data message is received.
// Care should be taken that this method does not block, since doing
// so would block the entire UDP packet receive loop. Additionally, the byte
// slice may be modified after the call returns, so it should be copied if needed.
func (sd *SyncerDelegate) NotifyMsg(msg []byte) {

	check := CompareJson(msg, Neighbour.NeighboursList{})

	switch check {
	case MESSAGE:
		//all incoming message as Message struct will be handeld
		Message_Handling(msg, sd)
	case ELECTION_EXPLORER_MESSAGE:
		//It will handle the explorer message
		election_explorer_message_handling(msg, sd)
	case ELECTION_ECHO_MESSAGE:
		//it handeld the echo message. echo message contains the answer for coordinator election message.
		echo_message_handling(msg, sd)
	case NEIGHBOUR_INFO_MESSAGE:
		//MasterNode recieve's neighbours from every node in the cluster afther any update occurred
		//afther recieved the message it will insert nodes and their neighbour's in to "NodesAndNeighbours" list
		neighbour_info_message_handling(sd, msg)
	}
}

func neighbour_info_message_handling(sd *SyncerDelegate, msg []byte) {
	if sd.Node.LocalNode().Name == "Master" {

		var receivedMsg Neighbour.NeighboursList
		err := json.Unmarshal(msg, &receivedMsg)
		errorAnd_Msg := Error_And_Msg{Err: err, Text: "Could not encode the NeighboursInfo message"}
		Check(errorAnd_Msg)

		sd.NodesNeighbour.AddNodesAndNeighbours(receivedMsg)
	}
}

func echo_message_handling(msg []byte, sd *SyncerDelegate) {
	fmt.Println("Echo Message")
	echo_message := new(Election.Echo)
	err := json.Unmarshal(msg, echo_message)
	_ = err
	fmt.Println("Echo Sender: ", echo_message.EchoSender)
	fmt.Println("--------------: ", echo_message.EchoSenderList)
	fmt.Println("Koordinator Id--------------: ", echo_message.Coordinator)

	if echo_message.Coordinator == sd.ElectionExplorer.M {
		//update the local echo sender list
		UpdateLocalEchoMessage(sd, echo_message)
		if sd.EchoMessage.EchoWaitedNum == sd.EchoMessage.EchoRecievedNum {
			if sd.LocalNode.Name != sd.ElectionExplorer.Initiator.Name {
				//When er von all seine Nachbarn bekommen hat, traegt er die sender auf die EchoSenderList ein.
				echo_message.EchoSenderList = sd.EchoMessage.EchoSenderList
				sd.SendMesgToMember(*sd.ElectionExplorer.Initiator, sd.EchoMessage)

			} else {
				fmt.Println("I am the Coordinator and i Recieved From : ", sd.EchoMessage.EchoSenderList)
				Inform_coordinator_success(*sd.EchoMessage, *sd)

			}
		}
	}
}

func Inform_coordinator_success(echo Election.Echo, sd SyncerDelegate) {
	msg := Message{Msg: "Iam_the_Coordinator", Snder: sd.Node.LocalNode().Name}

	for _, sender := range echo.EchoSenderList {
		sd.SendMesgToMember(*sender, msg)
	}

}

func election_explorer_message_handling(msg []byte, sd *SyncerDelegate) {
	fmt.Println("Explorerrrrrrrrrrrrrrrrrrrrrrrrrrrrrrr")
	explorer := new(Election.ElectionExplorer)
	err := json.Unmarshal(msg, explorer)
	_ = err

	fmt.Println("Message From: ", explorer.Initiator.Name, "**************: ", explorer.M)

	switch sd.CompaireElection(*explorer) {

	case RECIEVED_ExplorerID_EQUAL_TO_Local_ID:
		If_The_Same_Explorer_Recieved(sd, explorer)
		//if recieved explorer is greater than local explorer id and the neigbour list is greater than 1
	case RECIEVED_ExplorerID_GREATER_TO_Local_ID:

		switch sd.Check_If_Nod_is_Leaf() {
		case IS_NOT_LEAF_NODE:
			//th clean previous election process
			sd.EchoMessage.Clear()
			sd.ElectionExplorer.Clear()

			var tempExplorer Election.ElectionExplorer
			tempExplorer.Clear()
			tempExplorer = *explorer
			//init the local explorer struct with recieved to save all the info.
			sd.ElectionExplorer = &tempExplorer
			sd.ElectionExplorer.Add_RecievedFrom(*explorer.Initiator)
			explorer.Initiator = sd.LocalNode

			//hier wird bestimmt, dass der Node von alle seine Nachbarn echo Nachrichten erwartet
			//ausser der Sender
			sd.EchoMessage.EchoWaitedNum = len(sd.Neighbours.Neighbours) - 1

			sendExplorer(explorer, sd)
		case IS_LEAF_NODE:

			//send echo
			sd.ElectionExplorer.M = explorer.M
			sd.ElectionExplorer.Initiator = explorer.Initiator

			fmt.Println("Send Echo For the First time to : ", sd.ElectionExplorer.Initiator)

			sd.EchoMessage.Coordinator = sd.ElectionExplorer.M
			sd.EchoMessage.EchoSender = *sd.LocalNode
			sd.EchoMessage.AddSender(*sd.LocalNode)

			sd.SendMesgToMember(*sd.ElectionExplorer.Initiator, sd.EchoMessage)
		}

	}
}

//If the recieves the same explorer.
//if the sender is not containt in the local explorer sender list, then it will save to the local sender list and
//waited for echo var will decrement.
//If the recieved counter and waited counter are equal && the node is not the elction initiator,
//then node start to send echo to node, which recieved for the first time.
func If_The_Same_Explorer_Recieved(sd *SyncerDelegate, explorer *Election.ElectionExplorer) {
	if !sd.ElectionExplorer.ContainsNodeInRecievedFrom(explorer.Initiator) {
		//hier wird bestimmt, dass der Node von diesem Node auch keine Echo erwartet,
		//weil er der selbe Nachricht erhalten hat.
		sd.ElectionExplorer.Add_RecievedFrom(*explorer.Initiator)
		sd.EchoMessage.EchoWaitedNum--
		fmt.Println("$$$$$$$$$$$$$$$$$$$$$$$$$$$$4 Sender: ", explorer.Initiator, " ==========: ",
			sd.EchoMessage.EchoWaitedNum, "----mssage waited: ", sd.EchoMessage.EchoWaitedNum,
			" mssage recieved: ", sd.EchoMessage.EchoRecievedNum)
	}
	if sd.EchoMessage.EchoRecievedNum == sd.EchoMessage.EchoWaitedNum {
		if sd.ElectionExplorer.Initiator != sd.LocalNode {
			fmt.Println("Echo wird gesendet, Weil weiter Knoten in ein Zycklus sind")
			fmt.Println("Send Echo For the First time to : ", sd.ElectionExplorer.Initiator)

			sd.EchoMessage.Coordinator = sd.ElectionExplorer.M
			sd.EchoMessage.EchoSender = *sd.LocalNode
			sd.EchoMessage.AddSender(*sd.LocalNode)

			sd.SendMesgToMember(*sd.ElectionExplorer.Initiator, sd.EchoMessage)
		}
	}
}

func (sd *SyncerDelegate) CompaireElection(eN Election.ElectionExplorer) int {

	if sd.ElectionExplorer.M < eN.M {
		return RECIEVED_ExplorerID_GREATER_TO_Local_ID
	} else if sd.ElectionExplorer.M == eN.M {
		return RECIEVED_ExplorerID_EQUAL_TO_Local_ID
	}
	return Default
}

func (sd *SyncerDelegate) Check_If_Nod_is_Leaf() int {
	if len(sd.Neighbours.Neighbours) == 1 {
		return IS_LEAF_NODE
	}
	if len(sd.Neighbours.Neighbours) > 1 {
		return IS_NOT_LEAF_NODE
	}
	return Default
}

//as message the following action will be handeld:
//body=leave -> the process will exist
//body=readNeighbour -> the node will read its neighbour from file
//body=Start_Election -> the node will start elcetion process
func Message_Handling(msg []byte, sd *SyncerDelegate) {
	var receivedMsg Message
	err := json.Unmarshal(msg, &receivedMsg)
	_ = err
	switch receivedMsg.Msg {
	case "Iam_the_Coordinator":
		fmt.Println("Message : ", receivedMsg.Msg, "Coordinator: ", receivedMsg.Snder)
	case "leave":
		//Leave will kill the process
		//and the node will remove from Cluster-Memberlist
		sd.Leave()

	case "readNeighbour":
		fmt.Println("Readed .dot file -----------------------------------------------")
		read_neighbours_from_dot_file(sd, receivedMsg)

	case "Start_Election":
		fmt.Println("I Have to Start Election Process Message to become coordinator +++++++++++++++++++++++++")
		start_election(sd)
		println("Election started----------------: ", sd.ElectionExplorer.M)
	}
}

func start_election(sd *SyncerDelegate) {
	nodeId, _ := strconv.Atoi(ParseNodeId(sd.LocalNode.Name))
	tempExplorer := Election.NewElection(nodeId, *sd.LocalNode)

	sd.ElectionExplorer = tempExplorer

	sd.EchoMessage.EchoWaitedNum = len(sd.Neighbours.Neighbours)
	sendExplorer(tempExplorer, sd)
}

//clear the available neighbour list
//Read Graph from file
//add Nodes to Neighbourlist if there is a releastionship for this nod found
//Send the new neighbour list to MasterNode
func read_neighbours_from_dot_file(sd *SyncerDelegate, receivedMsg Message) {
	sd.neighbourFilePath = &receivedMsg.FilePath
	ReadNeighbourFromDot(sd)

	for _, ne := range sd.Neighbours.Neighbours {
		fmt.Println("Neigbous: ", ne.Name)
	}
}

//update the local echo sender list
func UpdateLocalEchoMessage(sd *SyncerDelegate, echo_message *Election.Echo) {
	sd.EchoMessage.EchoRecievedNum++
	if sd.ElectionExplorer.Initiator.Name != sd.LocalNode.Name {
		sd.EchoMessage.AddSender(*sd.LocalNode)
	}
	sd.EchoMessage.AddSender(echo_message.EchoSender)
	sd.EchoMessage.Coordinator = sd.ElectionExplorer.M
	sd.EchoMessage.EchoSender = *sd.LocalNode

	for _, sender := range echo_message.EchoSenderList {
		sd.EchoMessage.EchoSenderList[sender.Name] = sender
	}

}

//Echo wird nur an sender der ExplorerNachricht gesendet

func sendExplorer(explorer *Election.ElectionExplorer, sd *SyncerDelegate) {
	body, err := json.Marshal(explorer)
	error_and_msg := Error_And_Msg{Err: err, Text: "Encode the rumors faild!"}
	Check(error_and_msg)

	for _, neighbour := range sd.Neighbours.Neighbours {
		if neighbour.Name != sd.ElectionExplorer.Initiator.Name {
			fmt.Println("##########################3 Send To: ", neighbour.Name)
			sd.Node.SendBestEffort(&neighbour, body)
		}

	}
}

func ReadNeighbourFromDot(sd *SyncerDelegate) {
	if sd.neighbourFilePath != nil && *sd.neighbourFilePath != "" {

		g := Graph.NewDiGraph()
		g.ParseFileToGraph(*sd.neighbourFilePath)

		// if AddNodesToNeighbourList(g, sd) {
		AddNodesToNeighbourList(g, sd)
		body, _ := json.Marshal(sd.Neighbours)
		sd.Node.SendBestEffort(sd.MasterNode, body)
		// }
	}
}

func AddNodesToNeighbourList(g *Graph.Graph, sd *SyncerDelegate) bool {
	sd.Neighbours.ClearNeighbours()
	for _, node := range g.Nodes {
		if node.Name == sd.LocalNode.Name {

			neighbours := g.GetEdges(node.Name)
			if len(neighbours.Nodes) > 0 {

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
	// d.Neighbours.UpdateNeighbourList(*d.NeighbourNum, *d.NodeList)
	// body, _ := json.Marshal(d.Neighbours)
	// d.Node.SendBestEffort(d.MasterNode, body)

	log.Printf("notify join %s(%s:%d)", node.Name, node.Addr.To4().String(), node.Port)

	// fmt.Printf("%s Neigbour's: \n", d.Neighbours.Node.Name)
	// for _, p := range d.Neighbours.Neighbours {
	// 	fmt.Println("Neigbour---------: ", p.Name)
	// }

}

func (d *SyncerDelegate) NotifyLeave(node *memberlist.Node) {

	log.Printf("notify leave %s(%s:%d)", node.Name, node.Addr.To4().String(), node.Port)

	if d.LocalNode.Name == "Master" {
		d.NodesNeighbour.RemoveNodesNeighbours(*node)

	}

	d.NodeList.RemoveNode(node)
	if d.Neighbours.Contains(node) {
		// d.Neighbours.UpdateNeighbourList(*d.NeighbourNum, *d.NodeList)
		d.Neighbours.RemoveNeighbour(*node)
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

func (sd *SyncerDelegate) SendMesgToList(list map[string]memberlist.Node, value interface{}) {
	if len(list) <= 0 {
		return
	}

	body, err := json.Marshal(value)
	error_and_msg := Error_And_Msg{Err: err, Text: "Encode the Struct faild!"}
	Check(error_and_msg)

	for _, member := range list {
		sd.Node.SendBestEffort(&member, body)
	}
}

func (sd *SyncerDelegate) SendMesgToMember(node memberlist.Node, value interface{}) {
	body, err := json.Marshal(value)
	error_and_msg := Error_And_Msg{Err: err, Text: "Encode the Struct faild!"}
	Check(error_and_msg)

	sd.Node.SendBestEffort(&node, body)
	fmt.Println("Echo Send To: ", node.Name)

}
