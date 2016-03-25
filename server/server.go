package server

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"log"
	"net"
	"os/exec"
	"time"
	"net/http"

	"crypto/rand"

	"isucdc.com/keyescrow-server/escrow"

	zmq "github.com/pebbe/zmq3"
)

const (
	GITLABKEY       = "F3KyJyinpFVf1Vq2Dj5M"
	GITLAB_USER_URL = "http://git.team1.isucdc.com/api/v3/users?username=%v&private_token=%v"
	GITLAB_USER_KEY = "http://git.team1.isucdc.com/api/v3/users/%d/keys?private_token=%v"
)

type authinfo struct {
	Token    string
	User     string
	Issued   time.Time
	IsAdmin  bool
}

var authTable map[string]authinfo = make(map[string]authinfo)

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

			err := SetGitlabKey(d.User, pub)

			if (err != nil) {
				log.Println(err)
			}

			go func(username string) {
				hosts := []string{"runner-1", "runner-2", "www", "git", "keyescrow", "shell"}

				for i := range hosts {
					host := hosts[i]
					var out bytes.Buffer
					cmd := exec.Command("scp", pub, fmt.Sprintf("%v:%v", host, "/tmp/key.pub"))
					cmd.Stderr = &out
					err = cmd.Run()
					if err != nil {
						fmt.Println(out.String())
						log.Printf("Error scp: %v", err.Error())
						continue
					}

					log.Printf("Dispatching key to %v", host)

					cmd = exec.Command("ssh", host, "mkdir", "-p", fmt.Sprintf("/home/%v/.ssh", username))
					cmd.Stderr = &out
					err = cmd.Run()
					if err != nil {
						fmt.Println(out.String())
						log.Printf("Error mkdir: %v", err.Error())
						continue
					}

					cmd = exec.Command("ssh", host, fmt.Sprintf("cat /tmp/key.pub >> /home/%v/.ssh/authorized_keys", username))
					cmd.Stderr = &out
					err = cmd.Run()
					if err != nil {
						fmt.Println(out.String())
						log.Printf("Error cat: %v", err.Error())
						continue
					}

					cmd = exec.Command("ssh", host, "chown", "-R",
						username, fmt.Sprintf("/home/%v/.ssh", username))
					cmd.Stderr = &out
					err = cmd.Run()
					if err != nil {
						fmt.Println(out.String())
						log.Printf("Error chown: %v", err.Error())
						continue
					}

					cmd = exec.Command("ssh", host, "chmod", "600", fmt.Sprintf("/home/%v/.ssh/authorized_keys", username))
					cmd.Stderr = &out
					err = cmd.Run()
					if err != nil {
						fmt.Println(out.String())
						log.Printf("Error chmod: %v", err.Error())
						continue
					}
				}
			}(d.User)

			d.Send(s.Responder)
		case KeyRequest:
			kr := message.(KeyRequest)
			log.Printf("Got Key Request for: %v", kr.User)

			if validateAuthToken(kr.User, kr.Token) == false {
				ErrorMessage{Message: "Invalid token"}.Send(s.Responder)
				continue
			}

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
			log.Printf("Got Key Set Request for: %v", kr.User)

			if validateAuthToken(kr.User, kr.Token) == false {
				ErrorMessage{Message: "Invalid token"}.Send(s.Responder)
				continue
			}

			key := escrow.New(kr.User, kr.PubKey, kr.PrivKey)
			if err := key.Save(); err != nil {
				ErrorMessage{Message: err.Error()}.Send(s.Responder)
				continue
			}

			kreq := KeyRequest{
				User: kr.User,
			}
			kreq.Send(s.Responder)
		case AuthRequest:
			ar := message.(AuthRequest)
			log.Printf("Got Auth Request for: %v", ar.User)

			if escrow.AuthUser(ar.User, ar.Password) == false {
				ErrorMessage{Message: "Invalid.Username or password."}.Send(s.Responder)
				continue
			}

			authtoken := createAuthToken(ar.User, ar.Password)
			if authtoken == "" {
				ErrorMessage{Message: "Error creating token."}.Send(s.Responder)
				continue
			}

			areq := AuthResponse{
				User:  ar.User,
				Token: authtoken,
			}
			areq.Send(s.Responder)
		}
	}
}

func (s *Server) Close() {
	s.Responder.Close()
}

func SetGitlabKey(user string, pubkey string) error {
	key, err := escrow.FindUserKey(user)
	if err != nil {
		return err
	}

	resp, err := http.Get(fmt.Sprintf(GITLAB_USER_URL, user, GITLABKEY))
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		resp.Body.Close()
		return err
	}
	resp.Body.Close()

	var jsonObj map[string]string

	err = json.Unmarshal(body, jsonObj)
	if err != nil {
		return err
	}

	log.Println(jsonObj["id"])

	resp, err = http.PostForm(fmt.Sprintf(GITLAB_USER_KEY, jsonObj["id"], GITLABKEY), url.Values{"id": {jsonObj["id"]}, "title": {"ssh-rsa"}, "key": {key.PublicKey}})
	if err != nil {
		return err
	}

	body, err = ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return err
	}

	log.Println(body)

	return nil

	//cmd := exec.Command("sh", "addssh.sh", key.User, "ssh-rsa", key.PublicKey)
	//cmd.Run()
}

func createAuthToken(username string, password string) string {
	var buffer []byte
	buffer = make([]byte, 16)
	token_read, err := rand.Read(buffer)
	if token_read != 16 || err != nil {
		return ""
	}
	authtoken := hex.EncodeToString(buffer)

	authfield := authinfo{authtoken, username, time.Now(), escrow.IsAdmin(username, password)}
	authTable[authtoken] = authfield

	return authtoken
}

func validateAuthToken(username string, token string) bool {
	authfield := authTable[token];

	if time.Since(authfield.Issued) < (5 * time.Minute) {
		if authfield.User == username {
			return true
		} else {
			return authfield.IsAdmin
		}
	}

	return false
}
