package escrow

import (
	"fmt"
	"os"
	"path"
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
