package escrow

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

var Keydir = ""

type Key struct {
	User       string
	PublicKey  string
	PrivateKey string
}

func New(user string, pub string, priv string) *Key {
	return &Key{
		User:       user,
		PublicKey:  pub,
		PrivateKey: priv,
	}
}

func Open(user string) (*Key, error) {
	key := &Key{User: user}

	pub, err := os.Open(fmt.Sprintf("%v/%v/pubkey", Keydir, user))
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(pub)
	if err != nil {
		return nil, err
	}
	key.PublicKey = string(data)

	priv, err := os.Open(fmt.Sprintf("%v/%v/privkey", Keydir, user))
	if err != nil {
		return nil, err
	}
	data, err = ioutil.ReadAll(priv)
	if err != nil {
		return nil, err
	}
	key.PrivateKey = string(data)

	return key, nil
}

func (k Key) String() string {
	return fmt.Sprintf("{ User: %v, PubKey: %v, PrivKey: %v }",
		k.User, k.PublicKey, k.PrivateKey)
}

func (k *Key) GetKeyDir() string {
	return fmt.Sprintf("%v/%v/", Keydir, k.User)
}

func (k *Key) Save() error {
	log.Printf("Saving Key for User %v to: %v\n", k.User, k.GetKeyDir())
	if _, err := os.Stat(k.GetKeyDir()); os.IsNotExist(err) {
		if err = os.MkdirAll(k.GetKeyDir(), 0777); err != nil {
			log.Printf("There was an error saving the key: %v", err.Error())
			return err
		}
	}

	log.Printf("Saving public key")
	f, err := os.Create(k.GetKeyDir() + "pubkey")
	if err != nil {
		return err
	}

	_, err = f.WriteString(k.PublicKey)
	if err != nil {
		log.Printf("Could not write public key")
		return err
	}

	log.Printf("Saving private key")
	f, err = os.Create(k.GetKeyDir() + "privkey")
	if err != nil {
		return err
	}

	_, err = f.WriteString(k.PrivateKey)
	if err != nil {
		log.Printf("Could not write private key")
		return err
	}

	return nil
}

func (k *Key) Delete() error {
	return os.RemoveAll(k.GetKeyDir())
}
