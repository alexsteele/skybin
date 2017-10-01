package repo

import (
	"fmt"
	"io"
	"os"
	"path"
)

type PeerStore interface {
	StoreBlock(blockId string, block io.Reader) error
	GetBlock(blockId string) (block io.ReadCloser, err error)
}

type peerStore struct {
	peerdir string
}

func (ps *peerStore) StoreBlock(blockId string, block io.Reader) error {
	path := path.Join(ps.peerdir, blockId)

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return fmt.Errorf("Block %s already stored", blockId)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("Error storing block: %s", err)
	}
	defer f.Close()

	_, err = io.Copy(f, block)
	if err != nil {
		_ = os.Remove(path)
		return fmt.Errorf("Unable to write block: %s", err)
	}

	return nil
}

func (ps *peerStore) GetBlock(blockId string) (io.ReadCloser, error) {
	path := path.Join(ps.peerdir, blockId)

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("No block with id %s", blockId)
		}
		return nil, fmt.Errorf("Error reading block %s", blockId)
	}
	return f, nil
}
