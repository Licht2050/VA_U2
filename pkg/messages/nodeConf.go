package messages

import (
	"VAA_Uebung1/CSVutil"
	"VAA_Uebung1/pkg/Exception"
	"fmt"
	"math/rand"
)

/*Define Struct for Nodes in a Cluster*/
type NodeConf struct {
	all       []NodeInfo
	Self      *NodeInfo
	Neighbors []*NodeInfo
}

/*Configuration for Reading Rows from CSV-File*/
func NewNodeConf(csvName string, Id string) *NodeConf {
	config := &NodeConf{}
	for _, row := range CSVutil.ReadCSVRows(csvName) {
		config.all = append(config.all, NodeInfo{
			NodeId:     row[0],
			NodeIpAddr: row[1],
			Port:       row[2],
		})
	}
	self, err := config.find(Id)
	Exception.ErrorHandler(err)
	config.Self = self
	return config

}

/* Find Configuration for Specific Id*/
func (config *NodeConf) find(id string) (*NodeInfo, error) {
	for i := range config.all {
		if config.all[i].NodeId == id {
			return &config.all[i], nil
		}
	}
	return nil, fmt.Errorf("could not find configuration for entry with ID %s", id)
}

//TODO: Meet with Sayed
/*!!-- Need to be expand --!!*/
func (config *NodeConf) ChooseRandomNeighbor(n int) {
	for _, randIndex := range rand.Perm(len(config.all)) {
		other := &config.all[randIndex]
		if len(config.Neighbors) < n && other != config.Self {
			config.Neighbors = append(config.Neighbors, other)
		}
	}
}
