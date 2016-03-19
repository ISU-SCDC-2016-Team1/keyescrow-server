package main

import (
	"fmt"
	"log"
	"net"

	"github.com/voxelbrain/goptions"

	"isucdc.com/keyescrow-server/escrow"
	"isucdc.com/keyescrow-server/server"
)

const (
	PUBKEY_NAME  = "%v"
	PRIVKEY_NAME = "%v.priv"
)

func main() {
	options := struct {
		Host *net.TCPAddr  `goptions:"-s, --server"`
		Help goptions.Help `goptions:"-h, --help, description='Show this help'"`

		goptions.Verbs
		Server struct {
			KeyDir string `goptions:"-k, --keydir, description='Key Directory'"`
		} `goptions:"server"`
	}{ // Default values go here
		Host: &net.TCPAddr{
			IP:   net.ParseIP("localhost"),
			Port: 7654,
		},
	}
	//options.Host.IP = net.ParseIP(os.Getenv("KE_HOST"))
	fmt.Println("CDC Key Escrow Server v1")
	goptions.ParseAndFail(&options)

	if len(options.Verbs) <= 0 {
		fmt.Println("You must specify a verb")
		return
	}

	switch options.Verbs {
	case "server":
		if options.Server.KeyDir == "" {
			options.Server.KeyDir = "./keys"
		}
		//log.Printf("Using %v for key storage\n", options.KeyDir)
		escrow.Keydir = options.Server.KeyDir
		serv, err := server.New(options.Host)
		defer serv.Close()
		if err != nil {
			log.Fatalf("There was an error starting the server: %v", err.Error())
		}

		serv.Keydir = options.Server.KeyDir
		serv.Loop()
		break
	}
}
