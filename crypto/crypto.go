package crypto

import (
	"bytes"
	//	"io/ioutil"
	//	"encoding/base64"
	"errors"
	"golang.org/x/crypto/openpgp"
	"os"
)

type Crypto struct {
	contacts   map[string]openpgp.EntityList
	identities openpgp.EntityList
}

func New(prefix, password string) (*Crypto, error) {
	c := new(Crypto)
	c.contacts = make(map[string]openpgp.EntityList)

	c.loadKeys(prefix, password)

	return c, nil
}

func (c *Crypto) AddKey(nick string, key []byte) error {
	keyBuffer := bytes.NewBuffer(key)
	list, err := openpgp.ReadKeyRing(keyBuffer)
	if err != nil {
		return err
	}
	c.contacts[nick] = list
	return nil
}

func (c *Crypto) EncryptFor(nick string, message []byte) ([]byte, error) {
	recipients, ok := c.contacts[nick]
	if !ok {
		return []byte{}, errors.New("Could Not find contact '" + nick + "'")
	}
	cipherText := new(bytes.Buffer)
	w, err := openpgp.Encrypt(cipherText, recipients, c.identities[0], nil, nil)
	if err != nil {
		return []byte{}, err
	}

	_, err = w.Write(message)
	if err != nil {
		return []byte{}, err
	}

	err = w.Close()
	if err != nil {
		return []byte{}, err
	}

	return cipherText.Bytes(), nil
	//	return ioutil.ReadAll(cipherText)
}

func (c *Crypto) Decrypt(cipherText []byte) ([]byte, error) {
	/*	cipherBin := make([]byte, base64.StdEncoding.DecodedLen(len(cipherText)))
		_, err := base64.StdEncoding.Decode(cipherText, cipherBin)
			if err != nil {
				return []byte{}, err
			}
	*/
	md, err := openpgp.ReadMessage(bytes.NewBuffer(cipherText), c.identities, nil, nil)
	if err != nil {
		return []byte{}, err
	}
	message := new(bytes.Buffer)
	_, err = message.ReadFrom(md.UnverifiedBody)

	return message.Bytes(), err
}

func (c *Crypto) PublicKey() []byte {
	buf := new(bytes.Buffer)
	for _, key := range c.identities {
		key.Serialize(buf)
	}
	return buf.Bytes()
}

func (c *Crypto) loadKeys(prefix, password string) {
	contactsFileBuffer, _ := os.Open(prefix + "/private.key")
	defer contactsFileBuffer.Close()
	entityList, err := openpgp.ReadKeyRing(contactsFileBuffer)
	if err != nil {
		panic(err)
	}

	// unlock keys
	for _, entity := range entityList {
		entity.PrivateKey.Decrypt([]byte(password))
		for _, subkey := range entity.Subkeys {
			subkey.PrivateKey.Decrypt([]byte(password))
		}
	}
	c.identities = entityList
}
