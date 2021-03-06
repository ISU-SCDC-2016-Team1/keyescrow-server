package escrow

import (
	"fmt"
	"os"
	"path"
	"log"
	"strings"
	"net/http"
	"io/ioutil"
	"crypto/tls"
	"github.com/tonnerre/go-ldap"
)

func FindUserKey(name string) (*Key, error) {
	dir := fmt.Sprintf("%v/%v", Keydir, name)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, err
	}

	return Open(name)
}

func UserKeyPath(name string) (string, string) {
	return path.Join(Keydir, name, "pubkey"), path.Join(Keydir, name, "prikey")
}

func AuthUser(username string, password string) bool {
	ld, err := ldap.Dial("tcp", "10.4.4.2:389")
	if err != nil {
		log.Printf("Error LDAP Connect: %v\n", err);
		return false
	}

	ustring := fmt.Sprintf("cn=%v,cn=users,dc=team1,dc=isucdc,dc=com", username)

	err = ld.Bind(ustring, password)
	if err != nil {
		log.Printf("Error LDAP Bind (%v,%v): %v\n", ustring, password, err);
		return false
	}
	return true
}

func IsAdmin(username string, password string) bool {
	tr := &http.Transport{
        	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
    	}
	client := &http.Client{Transport: tr}
	resp, err := client.Get("https://ldap.team1.isucdc.com/isAdmin.ashx?user=" + username)

	if err != nil {
		return false
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	log.Println(username, ":", strings.TrimSpace(string(body)))

	if (strings.TrimSpace(string(body)) == "True") {
		return true
	}

	return false;
}
