package Neighbour

import (
	"fmt"
	"log"
	"math/rand"

	"github.com/hashicorp/memberlist"
)

type NodesAndNeighbours struct {
	NeigboursList []*NeighboursList
}

func (nodesN *NodesAndNeighbours) AddNodesAndNeighbours(nl NeighboursList) {
	if nl.Node.Name == "" {
		log.Println("Add Node and there Neighbours faild: param is empty")
		return
	}

	for _, node := range nodesN.NeigboursList {

		if node.Node.Name == nl.Node.Name {

			node.Neighbours = nil
			node.Neighbours = append(node.Neighbours, nl.Neighbours...)

			return
		}
	}
	nodesN.NeigboursList = append(nodesN.NeigboursList, &nl)
}

func (nodesN *NodesAndNeighbours) RemoveNodesNeighbours(node memberlist.Node) {
	if node.Name == "" {
		log.Println("Add Node and there Neighbours faild: param is empty")
		return
	}

	for index := 0; index < len(nodesN.NeigboursList); index++ {
		if nodesN.NeigboursList[index].Node.Name == node.Name {
			//first clear the Neibourslist of the node and then the node
			nodesN.NeigboursList[index].Neighbours = nil
			nodesN.NeigboursList = append(nodesN.NeigboursList[:index], nodesN.NeigboursList[index+1:]...)
			return
		}
	}
}

type NodesList struct {
	Nodes []memberlist.Node
}

func (nodeList *NodesList) Clone() NodesList {
	var nl NodesList
	nl.Nodes = append(nl.Nodes, nodeList.Nodes...)
	return nl
}

func (neighbourList *NeighboursList) Clone() NeighboursList {
	var nl NeighboursList
	nl.Node = neighbourList.Node

	nl.Neighbours = append(nl.Neighbours, neighbourList.Neighbours...)

	return nl
}

func (nodeL *NodesList) AddNode(node *memberlist.Node) {
	if node == nil {
		log.Println("AddNode to the NodeList is failed: param is empty!")
	}
	for _, n := range nodeL.Nodes {
		if n.Name == node.Name {
			return
		}
	}
	nodeL.Nodes = append(nodeL.Nodes, *node)
}

func (nodeL *NodesList) RemoveNode(node *memberlist.Node) {
	if node == nil {
		log.Println("Remove Node from the NodeList is failed: param is empty!")
	}
	for index := 0; index < len(nodeL.Nodes); index++ {
		if nodeL.Nodes[index].Name == node.Name {
			nodeL.Nodes = append(nodeL.Nodes[:index], nodeL.Nodes[index+1:]...)
			return
		}
	}
}

type NeighboursList struct {
	Node       memberlist.Node   `json:"node"`
	Neighbours []memberlist.Node `json:"neighbours"`
}

func (nl *NeighboursList) String() string {
	out := nl.Node.Name + " Neigbours {\n"
	for _, neigbour := range nl.Neighbours {
		out += fmt.Sprintf("\tName: %s\t Port: %d\n", neigbour.Name, neigbour.Port)
	}
	out += "}\n"

	return out
}

func (nl *NeighboursList) ChooseRandomNeighbour(num int, randNeighbour map[string]memberlist.Node) {
	if num > len(nl.Neighbours) {
		fmt.Println("Die Anzahl der RandMember soll kleine als Anzahl der Neighbour sein!")
		return
	}

	// temp := make(map[string]memberlist.Node)
	for i := 0; i < num; {

		rIndex := rand.Intn(len(nl.Neighbours))
		if _, ok := randNeighbour[nl.Neighbours[rIndex].Name]; !ok {
			randNeighbour[nl.Neighbours[rIndex].Name] = nl.Neighbours[rIndex]
			i++
		}

	}
}

func (nl *NeighboursList) Contains(node *memberlist.Node) bool {
	if node == nil {
		log.Println("Search the nod in the NeighbourList failed: param is empty!")
		return false
	}

	for _, neigbour := range nl.Neighbours {
		if neigbour.Name == node.Name {
			return true
		}
	}
	return false
}

func NewNeighbourList() *NeighboursList {
	return &NeighboursList{}
}

func (nl *NeighboursList) AddNeighbour(node memberlist.Node) bool {
	for _, neighbrour := range nl.Neighbours {
		if neighbrour.Name == node.Name {
			//if node in neigbourlist already exits!
			return false
		}
	}

	nl.Neighbours = append(nl.Neighbours, node)
	return true
}

func (nl *NeighboursList) ClearNeighbours() {
	nl.Neighbours = nil
}

func (nl *NeighboursList) RemoveNeighbour(node memberlist.Node) bool {
	if node.Name == "" {
		log.Println("Remove Node error: param is emtpy")
		return false
	}
	for index := 0; index < len(nl.Neighbours); {
		if nl.Neighbours[index].Name == node.Name {
			nl.Neighbours = append(nl.Neighbours[:index], nl.Neighbours[index+1:]...)
			return true
		}
		index++
	}

	return false
}

func (nl *NeighboursList) UpdateNeighbourList(neighbourNum int, list NodesList) {
	list_temp := list.Clone()
	//The Neighbours will chose every time new
	nl.Neighbours = nil

	if len(list_temp.Nodes) == 0 || neighbourNum > len(list_temp.Nodes) {
		log.Printf("NodesLis is empty or neighbourNum is < len(nodelist)!\n")
		return
	}
	//the for loop check the len of NeigboursList too.
	for i := 0; i < neighbourNum && len(nl.Neighbours) != neighbourNum; {
		//chose randomly a node
		rIndex := rand.Intn(len(list_temp.Nodes))
		if nl.AddNeighbour(list_temp.Nodes[rIndex]) {
			i++
		}
		//remove the node from list to make the chose func faster
		list_temp.Nodes = append(list_temp.Nodes[:rIndex], list_temp.Nodes[rIndex+1:]...)
	}
}
