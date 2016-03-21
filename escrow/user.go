package escrow

import (
	"fmt"
	"os"
	"path"

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
	ld, err := ldap.Dial("192.168.1.1", "387")
    if err != nil {
        return false
    }

	err = ld.Bind(username, password)
	if err == nil {
		return true
	}
	return false
}
