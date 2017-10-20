package peer

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	core "skybin/core/proto"
)

type Options struct {
	core.ProviderInfo
	Dir string // Dir is the directory where peers' content is stored.
}

func New(options Options) (core.Provider, error) {
	return &provider{
		Options: options,
	}, nil
}

// provider implements core.Provider
type provider struct {
	Options
}

func (p *provider) Info() (*core.ProviderInfo, error) {
	return &p.ProviderInfo, nil
}

func (p *provider) Negotiate(contract *core.Contract) (*core.Contract, error) {

	// TODO: Implement.
	c := *contract
	c.ProviderSignature = "sig"
	return &c, nil
}

func (p *provider) StoreBlock(ID string, block []byte) error {
	path := path.Join(p.Dir, ID)

	ioutil.WriteFile(path, block, 0666)

	return nil
}

func (p *provider) GetBlock(ID string) (block []byte, err error) {
	path := path.Join(p.Dir, ID)

	block, err = ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("No block with id %s", ID)
		}
		return nil, fmt.Errorf("Error reading block %s", ID)
	}
	return block, nil
}
