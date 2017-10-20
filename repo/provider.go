package repo

import (
	"encoding/json"
	"errors"
	"os"
	"path"
	core "skybin/core/proto"
)

func (r *repo) listProviders() ([]core.PeerInfo, error) {
	if r.pcache != nil {
		return r.pcache, nil
	}

	pvdrs, err := loadProviders(path.Join(r.homedir, "providers.json"))
	if err != nil {
		return nil, err
	}

	r.pcache = pvdrs
	return pvdrs, nil
}

func loadProviders(filename string) ([]core.PeerInfo, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var providers []core.PeerInfo
	err = json.NewDecoder(f).Decode(&providers)
	if err != nil {
		return nil, err
	}

	return providers, nil
}

func (r *repo) getProviderInfo(providerID string) (*core.PeerInfo, error) {
	pvdrs, err := r.listProviders()
	if err != nil {
		return nil, err
	}
	for _, pvdr := range pvdrs {
		if pvdr.ID == providerID {
			return &pvdr, nil
		}
	}
	return nil, errors.New("provider not found")
}
