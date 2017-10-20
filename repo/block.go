package repo

import (
	"crypto/sha1"
	"encoding/base32"

	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	core "skybin/core/proto"
)

func hash(data []byte) string {
	h := sha1.New()
	h.Write(data)
	return base32.StdEncoding.EncodeToString(h.Sum(nil))
}

func makeBlockId(ownerID string, name string) string {
	s := fmt.Sprintf("%s:%s", ownerID, name)
	return hash([]byte(s))
}

func loadINodeBlock(filename string) (*core.INodeBlock, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	block := &core.INodeBlock{}
	err = json.NewDecoder(f).Decode(block)
	return block, err
}

func loadDirBlock(filename string) (*core.DirBlock, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	block := &core.DirBlock{}
	err = json.NewDecoder(f).Decode(block)
	return block, err
}

func marshalBlock(block interface{}) ([]byte, error) {
	var buf bytes.Buffer
	e := json.NewEncoder(&buf)
	err := e.Encode(block)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func saveBlock(filename string, block interface{}) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(block)
}

type blockInfo struct {
	ID   string
	Size int
}

func getBlockInfo(file io.Reader, blockSize int) ([]blockInfo, error) {
	var info []blockInfo
	for {
		buf, err := readNextBlock(file, blockSize)
		if err != nil {
			return nil, err
		}
		if len(buf) == 0 {
			break
		}
		info = append(info, blockInfo{
			ID:   hash(buf),
			Size: len(buf),
		})
	}
	return info, nil
}

func readNextBlock(file io.Reader, blockSize int) ([]byte, error) {
	buf := make([]byte, blockSize)
	nr := 0
	for nr < len(buf) {
		n, err := file.Read(buf[nr:])
		if err != nil && err != io.EOF {
			return nil, err
		}
		if n == 0 {
			break
		}
		nr += n
	}
	return buf[:nr], nil
}
