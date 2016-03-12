package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os/exec"

	"isucdc.com/keyescrow/escrow"

	zmq "github.com/pebbe/zmq3"
)

const (
	GITLABKEY       = "UyrcEQzJwmoaEiTHtRjf"
	GITLAB_USER_URL = "http://gitlab/api/v3/users?username=%v&private_token=%v"
	GITLAB_USER_KEY = "http://gitlab/api/v3/users/%d/keys?private_token=%v"
)

type Server struct {
	Responder *zmq.Socket
	Keydir    string
}

func New(addr *net.TCPAddr) (*Server, error) {
	sock, err := zmq.NewSocket(zmq.REP)
	sock.Bind(fmt.Sprintf("tcp://%v", addr))
	log.Printf("Listening on %v", addr)
	return &Server{
		Responder: sock,
	}, err
}

func (s *Server) Loop() {
	log.Println("Starting loop")
	for {
		msg, _ := s.Responder.RecvBytes(0)
		message := RecvMsg(msg)

		switch message.(type) {
		case Dispatch:
			d := message.(Dispatch)
			log.Printf("Got Dispatch for %v", d.User)

			pub, _ := escrow.UserKeyPath(d.User)

			file, err := ioutil.ReadFile("./hosts.json")
			if err != nil {
				errmsg := "Could not read hosts file"
				log.Println(errmsg)
				ErrorMessage{Message: errmsg}.Send(s.Responder)
				continue
			}

			var jsontype interface{}
			json.Unmarshal(file, &jsontype)

			SetGitlabKey(d.User, pub)

			var hosts = jsontype.(map[string]interface{})
			for host, user := range hosts {
				var out bytes.Buffer
				cmd := exec.Command("scp", pub, fmt.Sprintf("%v@%v:/tmp/%v.pub",
					user, host, d.User))
				cmd.Stderr = &out
				err = cmd.Run()
				if err != nil {
					fmt.Println(out.String())
					log.Printf("Error scp: %v", err.Error())
				}

				hoststr := fmt.Sprintf("%v@%v", user, host)
				log.Printf("Dispatching key to %v", hoststr)

				cmd = exec.Command("ssh", hoststr, "mkdir", "-p", fmt.Sprintf("/home/%v/.ssh", d.User))
				cmd.Stderr = &out
				err = cmd.Run()
				if err != nil {
					fmt.Println(out.String())
					log.Printf("Error mkdir: %v", err.Error())
				}

				cmd = exec.Command("ssh", hoststr,
					fmt.Sprintf("cat /tmp/%v.pub >> /home/%v/.ssh/authorized_keys",
						d.User, d.User))
				cmd.Stderr = &out
				err = cmd.Run()
				if err != nil {
					fmt.Println(out.String())
					log.Printf("Error cat: %v", err.Error())
				}

				cmd = exec.Command("ssh", hoststr, "chown", "-R",
					d.User, fmt.Sprintf("/home/%v/.ssh", d.User))
				cmd.Stderr = &out
				err = cmd.Run()
				if err != nil {
					fmt.Println(out.String())
					log.Printf("Error chown: %v", err.Error())
				}

				cmd = exec.Command("ssh", hoststr, "chmod", "600", fmt.Sprintf("/home/%v/.ssh/authorized_keys", d.User))
				cmd.Stderr = &out
				err = cmd.Run()
				if err != nil {
					fmt.Println(out.String())
					log.Printf("Error chmod: %v", err.Error())
				}
			}

			d.Send(s.Responder)
		case KeyRequest:
			kr := message.(KeyRequest)
			log.Printf("Got Key Request for: %v", kr.User)

			key, err := escrow.FindUserKey(kr.User)
			if err != nil || key == nil {
				errmsg := fmt.Sprintf("Could not find key for %v", kr.User)
				log.Println(errmsg)
				ErrorMessage{Message: errmsg}.Send(s.Responder)
				continue
			}

			kresp := KeyResponse{
				User:    kr.User,
				PubKey:  key.PublicKey,
				PrivKey: key.PrivateKey,
			}

			kresp.Send(s.Responder)
		case KeyResponse:
			kr := message.(KeyResponse)
			log.Printf("Get Key Set Request for: %v", kr.User)

			key := escrow.New(kr.User, kr.PubKey, kr.PrivKey)
			if err := key.Save(); err != nil {
				ErrorMessage{Message: err.Error()}.Send(s.Responder)
				continue
			}

			// Backup the generated key
			cmd := exec.Command("cat", kr.PubKey, ">>", "/root/.ssh/authorized_keys")
			cmd.Run()

			kreq := KeyRequest{
				User: kr.User,
			}
			kreq.Send(s.Responder)
		}
	}
}

func (s *Server) Close() {
	s.Responder.Close()
}

func SetGitlabKey(user string, pubkey string) {
	key, err := escrow.FindUserKey(user)
	if err != nil {
		return
	}

	cmd := exec.Command("sh", "addssh.sh", key.User, "ssh-rsa", key.PublicKey)
	cmd.Run()
}
