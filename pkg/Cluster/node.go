package Cluster

import (
	"VAA_Uebung1/pkg/Neighbour"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/hashicorp/memberlist"
)

type (
	Node struct {
		Memberlist *memberlist.Memberlist
		Neigbour   *Neighbour.NeighboursList
	}
)

type Item struct {
	Ip     string `json:"ip"`
	Status string `json:"status"`
}

func (node *Node) handler(w http.ResponseWriter, req *http.Request) {

	var items []Item

	for _, member := range node.Memberlist.Members() {
		hostName := member.Addr.String()
		portNum := int(member.Port)
		seconds := 5
		timeOut := time.Duration(seconds) * time.Second
		conn, err := net.DialTimeout("tcp", hostName+":"+strconv.Itoa(portNum), timeOut)
		_ = conn
		if err != nil {
			items = append(items, Item{Ip: hostName + ":" + strconv.Itoa(portNum), Status: "DOWN"})
		} else {
			items = append(items, Item{Ip: hostName + ":" + strconv.Itoa(portNum), Status: "Hello World"})
		}
	}
	for _, neigbour := range node.Neigbour.Neighbours {
		log.Println("Test: ", neigbour.Name)
		// items = append(items, Item{Ip: "Test" + ":" + strconv.Itoa(8006), Status: neigbour.Adress})
	}

	js, err := json.Marshal(items)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)

}
