package Cluster

import (
	"VAA_Uebung1/pkg/Exception"
	"VAA_Uebung1/pkg/Graph"
	"VAA_Uebung1/pkg/Neighbour"

	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/hashicorp/memberlist"
)

var file = "mygraph.dot"

/**Bei Initieren einer Cluster wird:
 *Ein TCP-Port mit der Hilfe von memberlist "PKG" geoeffnet
 *der Prozess bekommt: Id, BindAddr(IP) und BindPort.
 *Es wird SecreteKey generiert, der Hilft nur Prozess mit gleiche Key in der Liste hinzufuegen.
 *Es wird ein Liste erstellt.
 *Der generierte Prozess wird als Master-Node in der Liste hinzugefuegt.
 *

 */

func InitCluster(nodeName, bindIP, bindPort, httpPort string) {

	clusterKey := make([]byte, 32)
	_, err := rand.Read(clusterKey)
	if err != nil {
		Exception.ErrorHandler(err)
	}

	config := memberlist.DefaultLocalConfig()
	config.Name = nodeName
	config.BindAddr = bindIP
	config.BindPort, _ = strconv.Atoi(bindPort)
	config.SecretKey = clusterKey
	config.LogOutput = ioutil.Discard

	ml, err := memberlist.Create(config)
	err_st := Error_And_Msg{Err: err}
	Check(err_st)

	g := Graph.NewDiGraph()

	neigbours := Neighbour.NewNeighbourList()
	neigbours.Node = *ml.LocalNode()
	neigbourNum := 3
	nodeList := new(Neighbour.NodesList)
	nodesNeighbour := Neighbour.NodesAndNeighbours{}

	AddClusterMemberToNodeList(ml, nodeList)

	rumors_list := NewRumorsList()
	blievableRRNum := 2

	sd := &SyncerDelegate{
		Node: ml, Neighbours: neigbours, NeighbourNum: &neigbourNum,
		NodeList: nodeList, Graph: g,
		LocalNode: ml.LocalNode(), MasterNode: ml.LocalNode(),
		NodesNeighbour:       &nodesNeighbour,
		neighbourFilePath:    &file,
		RumorsList:           rumors_list,
		BelievableRumorsRNum: &blievableRRNum,
	}

	config.Delegate = sd
	config.Events = sd

	node := Node{
		Memberlist: ml,
	}

	log.Printf("new cluster created. key: %s\n", base64.StdEncoding.EncodeToString(clusterKey))
	//write the key in to file
	write_key_to_file(base64.StdEncoding.EncodeToString(clusterKey), int(ml.LocalNode().Port), "clusterKey.txt")

	http.HandleFunc("/", node.handler)

	// go func(){
	// 	tcp.ListenAndServe(":" + )
	// }()

	go func() {
		http.ListenAndServe(":"+httpPort, nil)
	}()

	log.Printf("webserver is up. URL: http://%s:%s/ \n", bindIP, httpPort)

	// time.Sleep(time.Second * 10)

	for i := 0; i < 7; i++ {
		Menu()
		userInput(ml, g, neigbours, &nodesNeighbour, *sd, rumors_list)
	}

	// for {
	// 	for _, memb := range ml.Members() {
	// 		println("Test----------------: %s", memb.Name)
	// 		time.Sleep(1 * time.Second)
	// 	}
	// }

	incomingSigs := make(chan os.Signal, 1)
	signal.Notify(incomingSigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, os.Interrupt)

	go func() {
		<-incomingSigs
		err := ml.Leave(time.Second * 5)
		err_st.Err = err
		Check(err_st)
	}()

	// select {
	// case <-incomingSigs:
	// 	fmt.Printf("Hello")
	// 	err := ml.Leave(time.Second * 5)
	// 	err_st.Err = err
	// 	Check(err_st)
	// }
}

func print_all_nodes(ml *memberlist.Memberlist) {
	for _, memb := range ml.Members() {

		fmt.Println(memb.Name)

	}
}

func stop_node(ml *memberlist.Memberlist, nodeId string) {
	msg := Message{Msg: "leave"}
	for _, memb := range ml.Members() {
		if memb.Name == nodeId {
			body, _ := json.Marshal(msg)
			ml.SendBestEffort(memb, body)
			println("Successfully stoped----------------: ", memb.Name)
		}
	}
}

func userInput(ml *memberlist.Memberlist, g *Graph.Graph,
	n *Neighbour.NeighboursList,
	nodesNeighbour *Neighbour.NodesAndNeighbours,
	sd SyncerDelegate, rumorslist *RumorsList) {

	input_value := 0
	UserInputInt(&input_value)

	menuInput := MenuEnum(input_value)

	switch menuInput {
	case Kill_Node:
		node_id := ""
		fmt.Println("please insert the node-id")
		fmt.Scanf("%s", &node_id)
		stop_node(ml, node_id)
	case Kill_AllNodes:
		msg := Message{Msg: "leave"}
		BroadcastClusterMessage(ml, &msg)
	case Print_Cluster_Nodes:
		print_all_nodes(ml)
	case Send_Rumor:
		SendRumors(ml, rumorslist, &sd)
	case ParseNeighbourG_To_PNG:

		parseDiGToPNG(g, nodesNeighbour)
	case Print_Cluster_Nodes_Neighbours:
		print_all_neigbour(nodesNeighbour)
	case Print_Cluster_Nodes_Neighbours_As_DiG:
		print_neighbours_as_digraph(g, nodesNeighbour)
	case Pars_DiG_To_Dot:
		parseDiGToDot(g, nodesNeighbour)
	case Read_DiG_From_DotFile:
		readGraphFromFile()
	case Read_Neighbours_From_DotFile:
		readNeighbours_from_file(ml, sd)
	case Create_Rondom_Graph:
		nodeNum := -1
		edgeNum := -1
		Input(&nodeNum, &edgeNum)
		g := Graph.RondomDiGraph(nodeNum, edgeNum)
		fmt.Println(g.String())
		// g.ParseGraphToFile("testGraph")
		// fmt.Println("Successfully added to the file testgraph.dot")
	}

}

func readNeighbours_from_file(ml *memberlist.Memberlist, sd SyncerDelegate) {
	msg := Message{Msg: "readNeighbour"}
	msg.FilePath = "mygraph.dot"
	BroadcastClusterMessage(ml, &msg)
	sd.Node.UpdateNode(time.Millisecond * 2)

	println("All Nodes has updated their neighbour's list----------------")
}

func print_all_neigbour(nn *Neighbour.NodesAndNeighbours) {

	for _, neighbour := range nn.NeigboursList {
		fmt.Println(neighbour.String())
	}
}

func parseNodeNeighbours_to_digraph(g *Graph.Graph, nn *Neighbour.NodesAndNeighbours) {

	for _, n := range nn.NeigboursList {
		g.AddNode(n.Node.Name)
		for _, neighbour := range n.Neighbours {
			g.AddNode(neighbour.Name)
			g.AddEdge(n.Node.Name, neighbour.Name)
		}
	}
}

func print_neighbours_as_digraph(g *Graph.Graph, nn *Neighbour.NodesAndNeighbours) {
	//Initiate the graph var
	g.Clear()
	parseNodeNeighbours_to_digraph(g, nn)

	fmt.Println(g.String())
	fmt.Println("End of Neigbour-------------------")
}

func readGraphFromFile() {
	path := ""
	fmt.Printf("Enter file name \".dot\": ")
	fmt.Scanf("%s", &path)
	g := Graph.NewDiGraph()
	g.ParseFileToGraph(path)
	fmt.Println(g.String())
}

func parseDiGToDot(g *Graph.Graph, nn *Neighbour.NodesAndNeighbours) {
	file_name := ""
	fmt.Printf("Enter file name without \".dot\": ")
	fmt.Scanf("%s", &file_name)

	print_neighbours_as_digraph(g, nn)
	g.ParseGraphToFile(file_name)
}

func parseDiGToPNG(g *Graph.Graph, nn *Neighbour.NodesAndNeighbours) {
	file_name := ""
	fmt.Printf("Enter file name without \".png\": ")
	fmt.Scanf("%s", &file_name)

	print_neighbours_as_digraph(g, nn)
	g.ParseGraphToPNGFile(file_name)
}

func SendRumors(ml *memberlist.Memberlist, rumorsList *RumorsList, sd *SyncerDelegate) {
	input := -1
	fmt.Println("Empfang von wie viel knotes ist der Geruecht glaubhaft: ")
	UserInputInt(&input)
	sd.BelievableRumorsRNum = &input

	msg := Message{Msg: "HTW wid im Sommersemester 2022 auf die Onlinevorlesung umgestellt"}
	rumors := new(Rumors)
	rumors.RummorsMsg = msg

	rumorsList.AddRumorsToList(*rumors)
	sd.SendMsgToNeighbours(rumors)

	println("Gereucht initiert----------------: %s", ml.LocalNode().Name)
	time.Sleep(1 * time.Second)
}

func Input(nodeNum *int, edgeNum *int) {
	fmt.Println("Nodeanzahl eingeben: ")
	UserInputInt(nodeNum)

	fmt.Println("Edgesanzah eingeben: ")
	for {
		UserInputInt(edgeNum)
		if *edgeNum >= *nodeNum {
			break
		}
		fmt.Println("Edgesanzahl >= Nodeanzahl eingeben: ")
	}
}
