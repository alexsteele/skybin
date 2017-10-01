package repo

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base32"
	"encoding/json"
	"encoding/pem"
	"io/ioutil"
	"log"
	"os"
	"path"
	"skybin/core"
	"time"
)

func hash(data []byte) string {
	h := sha1.New()
	h.Write(data)
	return base32.StdEncoding.EncodeToString(h.Sum(nil))
}

func Init(homedir string) {

	if _, err := os.Stat(homedir); err == nil {
		log.Fatalf("error: %s already exists", homedir)
	}

	// Create repo directories
	checkErr(os.MkdirAll(homedir, 0700))
	checkErr(os.MkdirAll(path.Join(homedir, "keys"), 0700))
	checkErr(os.MkdirAll(path.Join(homedir, "peer"), 0700))
	checkErr(os.MkdirAll(path.Join(homedir, "user"), 0700))

	// Create server keys
	serverkey, err := rsa.GenerateKey(rand.Reader, 2048)
	checkErr(err)
	savePrivateKey(serverkey, path.Join(homedir, "keys", "nodeid"))
	savePublicKey(serverkey.PublicKey, path.Join(homedir, "keys", "nodeid.pub"))

	// Create user keys
	userkey, err := rsa.GenerateKey(rand.Reader, 2048)
	checkErr(err)
	savePrivateKey(userkey, path.Join(homedir, "keys", "userid"))
	savePublicKey(userkey.PublicKey, path.Join(homedir, "keys", "userid.pub"))

	// Create user ID
	bytes, err := asn1.Marshal(userkey.PublicKey)
	checkErr(err)
	userId := hash(bytes)

	// Create and save default repo config file
	config := Config{
		UserId:     userId,
		DhtAddress: "0.0.0.0:8001",
		ApiAddress: "0.0.0.0:8002",
	}
	configBytes, err := json.MarshalIndent(&config, "", "    ")
	checkErr(err)
	checkErr(ioutil.WriteFile(path.Join(homedir, "config.json"), configBytes, 0666))

	// Create and save the user's root block
	rootBlock := core.MetaBlock{
		ID:           makeMetaId(userId, "/"),
		Name:         "/",
		LastModified: time.Now(),
		OwnerID:      userId,
	}

	checkErr(saveMetaBlock(homedir, rootBlock.ID, &rootBlock))
}

func savePrivateKey(key *rsa.PrivateKey, path string) {
	keyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}
	savePemBlock(keyBlock, path)
}

func savePublicKey(key rsa.PublicKey, path string) {
	bytes, err := asn1.Marshal(key)
	checkErr(err)
	keyBlock := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: bytes,
	}
	savePemBlock(keyBlock, path)
}

func savePemBlock(block *pem.Block, path string) {
	f, err := os.Create(path)
	checkErr(err)
	defer f.Close()
	checkErr(pem.Encode(f, block))
}

func checkErr(err error) {
	if err != nil {
		log.Fatal("init failure: ", err)
	}
}
