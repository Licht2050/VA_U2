package Graph

import (
	"VAA_Uebung1/pkg/Exception"
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type Node struct {
	Name  string
	Nodes map[string]*Node
}

func NewNode(node string) *Node {
	return &Node{
		Name:  node,
		Nodes: map[string]*Node{},
	}
}

func (g *Graph) AddNode(node string) {
	//if node already exist; then return
	if _, ok := g.Nodes[node]; ok {
		return
	}
	node1 := NewNode(node)
	g.Nodes[node] = node1
}

func (g *Graph) RemoveSelectedEdge(fromNode, removeNode string) {

	if _, ok := g.Nodes[fromNode]; ok {
		delete(g.Nodes[fromNode].Nodes, removeNode)
	}
}

func (g *Graph) RemoveEdge(node string) {
	if node == "" {
		log.Println("Remove Edge failed: param is empty!")
		return
	}
	if !g.Directed {
		g.Nodes[node].Nodes = make(map[string]*Node)
		// delete(g.Nodes, node)
	} else {
		for edgeNode := range g.Nodes[node].Nodes {
			g.RemoveSelectedEdge(edgeNode, node)
		}
	}
}

func (g *Graph) RemoveNode(node string) {
	if node == "" {
		log.Println("Remove node failed: param is empty!")
		return
	}

	if _, ok := g.Nodes[node]; ok {
		//first remove all edges
		g.RemoveEdge(node)
		delete(g.Nodes, node)
	}
}

type Graph struct {
	Directed bool
	Nodes    map[string]*Node
}

func NewDiGraph() *Graph {
	return &Graph{
		Directed: true,
		Nodes:    map[string]*Node{},
	}
}

func NewUnDiGraph() *Graph {
	return &Graph{
		Nodes: map[string]*Node{},
	}
}

func (g *Graph) Clear() {
	g.Nodes = make(map[string]*Node)
}

func (g *Graph) AddEdge(nodeFrom, nodeTo string) {
	node1 := g.Nodes[nodeFrom]
	node2 := g.Nodes[nodeTo]

	if node1 == nil || node2 == nil {
		panic("Nodes are not exist")
	}

	if _, ok := node1.Nodes[node2.Name]; ok {
		return
	}

	node1.Nodes[node2.Name] = node2
	g.Nodes[node1.Name] = node1

	if !g.Directed && node1.Name != node2.Name {
		node2.Nodes[node1.Name] = node1
		g.Nodes[node2.Name] = node2
	}

}

func (g *Graph) GetEdges(node string) *Node {
	return g.Nodes[node]
}

func (g *Graph) String() string {
	//if Graph is directed
	if g.Directed {
		out := `digraph ClusterNodes {
			graph [ dpi = 600 ]; 
			rankdir=UD;
			size="8,5";
			node [shape = circle];`
		out += "\n"
		for k := range g.Nodes {
			for _, v := range g.GetEdges(k).Nodes {
				out += fmt.Sprintf("\t%s -> %s\n", k, v.Name)
			}
		}
		out += "}"
		return out
	}

	out := `graph ClusterNodes {
		graph [ dpi = 600 ]; 
		rankdir=UD;
		size="8,5";
		node [shape = circle];`
	out += "\n"
	var temp = make(map[string]string)
	for k := range g.Nodes {
		temp[k] = k
		for _, v := range g.GetEdges(k).Nodes {
			if _, ok := temp[v.Name]; !ok {
				out += fmt.Sprintf("\t%s -- %s\n", k, v.Name)
			}
		}
	}
	out += "}"
	return out
}

func (g *Graph) ParseStringToDiG(data string) {
	if len(strings.TrimSpace(data)) == 0 {
		panic("Variable is empty!")
	}

	scanner := bufio.NewScanner(strings.NewReader(data))
	//it read the first line and validate if it is directed graph
	scanner.Scan()
	graph := strings.Fields(scanner.Text())
	if graph[0] == "digraph" {
		for scanner.Scan() {
			line := scanner.Text()
			str := regexp.MustCompile(`[\w\d]+.\->.[\w\d]+`)

			for _, element := range str.FindAllString(line, -1) {
				g.AddNode(strings.TrimSpace(strings.Split(element, "->")[0]))
				g.AddNode(strings.TrimSpace(strings.Split(element, "->")[1]))
				g.AddEdge(strings.TrimSpace(strings.Split(element, "->")[0]),
					strings.TrimSpace(strings.Split(element, "->")[1]),
				)
			}
		}
		return
	}
	//if graph is undirected
	g.Directed = false
	for scanner.Scan() {
		line := scanner.Text()
		str := regexp.MustCompile(`[\w\d]+.\--.[\w\d]+`)

		for _, element := range str.FindAllString(line, -1) {
			g.AddNode(strings.TrimSpace(strings.Split(element, "--")[0]))
			g.AddNode(strings.TrimSpace(strings.Split(element, "--")[1]))
			g.AddEdge(strings.TrimSpace(strings.Split(element, "--")[0]),
				strings.TrimSpace(strings.Split(element, "--")[1]),
			)
			g.AddEdge(strings.TrimSpace(strings.Split(element, "--")[1]),
				strings.TrimSpace(strings.Split(element, "--")[0]),
			)
		}
	}

}

func (g *Graph) ParseFileToGraph(path string) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	g.ParseStringToDiG(string(data))
}

//param: filename without ".dot"
func (g *Graph) ParseGraphToFile(filename string) {
	f, err := os.Create(filename + ".dot")
	if err != nil {
		Exception.ErrorHandler(err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	_, err = fmt.Fprintf(w, "%s\n", g.String())
	if err != nil {
		Exception.ErrorHandler(err)
	}

	w.Flush()
}

func (g *Graph) ParseGraphToPNGFile(filename string) {
	cmd := exec.Command("dot", "-Tpng", "-o"+filename+".png")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	asDot := g.String()
	stdin.Write([]byte(asDot))
	stdin.Close()

	fmt.Println("Parsing graph to PNG file: ", filename+".png")
	cmd.Wait()
}

func RondomGraph(nodeNum, edgeNum int, directed bool) Graph {
	var g *Graph
	if directed {
		g = NewDiGraph()
	} else {
		g = NewUnDiGraph()
	}

	for index := 0; index < nodeNum; index++ {
		g.AddNode("Philisoph" + strconv.Itoa(index))
	}

	randNode := ""
	for _, node := range g.Nodes {
		for {
			randNode = ChoseRondomFromMap(g)
			if node.Name != randNode {
				break
			}
		}
		g.AddEdge(node.Name, randNode)
	}

	nodeRight := ""
	nodeLeft := ""
	for index := 0; index < edgeNum-len(g.Nodes); index++ {

		//if the radomly chosen edge already exists, do so until a non-existent combination is selected
		for {
			nodeRight = ChoseRondomFromMap(g)
			nodeLeft = ChoseRondomFromMap(g)

			if _, ok := g.Nodes[nodeLeft].Nodes[nodeRight]; !ok && nodeLeft != nodeRight {
				break
			}

		}
		g.AddEdge(nodeLeft, nodeRight)
	}
	return *g
}

func ChoseRondomFromMap(mp *Graph) string {
	temp_slice := []string{}
	for _, element := range mp.Nodes {
		temp_slice = append(temp_slice, element.Name)
	}

	randIndex := rand.Intn(len(temp_slice))
	return temp_slice[randIndex]
}
