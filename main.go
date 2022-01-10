package main

import (
	"VAA_Uebung1/pkg/Cluster"
	"flag"
	"fmt"
	"os"
)

func main() {

	joinCmd := flag.NewFlagSet("join", flag.ExitOnError)
	joinClusterKey := joinCmd.String("cluster-key", "", "cluster-key")
	joinNodeName := joinCmd.String("node-name", "", "node-name")
	joinKnownIP := joinCmd.String("known-ip", "", "known-ip")
	joinBindIP := joinCmd.String("bind-ip", "127.0.0.1", "bind-ip")
	joinBindPort := joinCmd.String("bind-port", "", "bind-port")
	joinHttpPort := joinCmd.String("http-port", "8888", "http-port")

	initCmd := flag.NewFlagSet("init", flag.ExitOnError)
	initNodeName := initCmd.String("node-name", "Master", "node-name")
	initBindIP := initCmd.String("bind-ip", "127.0.0.1", "bind-ip")
	initBindPort := initCmd.String("bind-port", "", "bind-port")
	initHttpPort := initCmd.String("http-port", "8888", "http-port")

	if len(os.Args) < 2 {
		fmt.Println("expected 'join' or 'init' subcommands")
		os.Exit(1)
	}

	switch os.Args[1] {

	case "join":
		joinCmd.Parse(os.Args[2:])

		Cluster.JoinCluster(*joinNodeName, *joinBindIP, *joinBindPort, *joinHttpPort, *joinClusterKey, *joinKnownIP)
	case "init":
		initCmd.Parse(os.Args[2:])

		Cluster.InitCluster(*initNodeName, *initBindIP, *initBindPort, *initHttpPort)
	default:
		fmt.Println("expected 'join' or 'init' subcommands")
		os.Exit(1)
	}

	os.Exit(0)

}
