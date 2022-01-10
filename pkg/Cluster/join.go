package Cluster

import (
	"VAA_Uebung1/pkg/Neighbour"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/hashicorp/memberlist"
)

func JoinCluster(nodeName, bindIP, bindPort, httpPort, clusterKey, knownIP string) {

	config := memberlist.DefaultLocalConfig()
	bIP, _ := strconv.Atoi(bindPort)
	config.BindPort = bIP
	// config.AdvertisePort = 8000
	config.BindAddr = bindIP
	config.Name = nodeName

	config.SecretKey, _ = base64.StdEncoding.DecodeString(clusterKey)
	//dadurch wird alle logging vo Memberlist ausgeschaltet
	config.LogOutput = ioutil.Discard

	ml, err := memberlist.Create(config)
	err_st := Error_And_Msg{Err: err}
	Check(err_st)

	_, err = ml.Join([]string{knownIP})
	//if err not nil print the following Text
	err_st.Err = err
	err_st.Text = "Failed to join cluster: "
	Check(err_st)

	log.Printf("Joined the cluster")

	// var broadcast *memberlist.TransmitLimitedQueue

	masterNode := SearchMemberbyName("Master", ml)

	neigbours := Neighbour.NewNeighbourList()
	neigbours.Node = *ml.LocalNode()
	//It defines the number of neighbours that each node can have
	neigbourNum := 3
	nodeList := new(Neighbour.NodesList)

	AddClusterMemberToNodeList(ml, nodeList)
	neigbours.UpdateNeighbourList(neigbourNum, *nodeList)
	body, _ := json.Marshal(neigbours)
	ml.SendBestEffort(masterNode, body)

	rumors_list := NewRumorsList()
	blievableRRNum := 2

	sd := &SyncerDelegate{
		Node: ml, Neighbours: neigbours, NeighbourNum: &neigbourNum,
		NodeList: nodeList, MasterNode: masterNode,
		LocalNode:            ml.LocalNode(),
		RumorsList:           rumors_list,
		BelievableRumorsRNum: &blievableRRNum,
	}

	config.Delegate = sd
	config.Events = sd

	node := Node{
		Memberlist: ml,
		Neigbour:   neigbours,
	}

	http.HandleFunc("/", node.handler)

	go func() {
		http.ListenAndServe(":"+httpPort, nil)
	}()

	log.Printf("webserver is up. URL: http://%s:%s/ \n", bindIP, httpPort)

	msg := Message{Msg: "I am a new Member", Snder: ml.LocalNode().Name, SendTime: time.Now()}
	time.Sleep(time.Second * 2)
	BroadcastClusterMessage(ml, &msg)

	incomingSigs := make(chan os.Signal, 1)
	signal.Notify(incomingSigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, os.Interrupt)

	select {
	case <-incomingSigs:
		log.Println("test")

		if err := ml.Leave(time.Second * 5); err != nil {
			err_st.Err = err
			Check(err_st)
		}
	case <-incomingSigs:

	}

}
