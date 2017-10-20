package repo

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path"
	core "skybin/core/proto"
	provider "skybin/provider/remote"
)

func DefaultHomeDir() (string, error) {
	user, err := user.Current()
	if err != nil {
		return "", err
	}
	return path.Join(user.HomeDir, ".skybin"), nil
}

type Info struct {
	HomeDir string
	Config  *Config
}

type Repo interface {
	Info() Info
	Put(filename string, opts *StorageOptions) error
	ListFiles() ([]string, error)
	Get(filename string, out io.Writer) error
	Sync() error
}

type repo struct {
	homedir   string
	config    *Config
	rootBlock *core.DirBlock
	pcache    []core.PeerInfo // Known storage providers
	logger    *log.Logger
}

func Open() (Repo, error) {
	homedir, err := findHomeDir()
	if err != nil {
		return nil, err
	}
	return OpenAt(homedir)
}

func OpenAt(homedir string) (Repo, error) {

	err := checkRepo(homedir)
	if err != nil {
		return nil, err
	}

	config, err := loadConfig(path.Join(homedir, "config.json"))
	if err != nil {
		return nil, err
	}

	rootBlock, err := loadDirBlock(path.Join(homedir, "user", makeBlockId(config.UserId, "/")))
	if err != nil {
		return nil, fmt.Errorf("Cannot load user's root block: %s", err)
	}

	logger := log.New(ioutil.Discard, "", log.Ldate|log.Ltime)
	if config.LogEnabled && len(config.LogFolder) > 0 {
		f, err := os.OpenFile(config.LogFolder, os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}
		logger.SetOutput(f)
	} else if config.LogEnabled {
		logger.SetOutput(os.Stdout)
	}

	return &repo{
		homedir:   homedir,
		config:    config,
		rootBlock: rootBlock,
		pcache:    nil,
		logger:    logger,
	}, nil
}

func (r *repo) Info() Info {
	return Info{
		HomeDir: r.homedir,
		Config:  r.config,
	}
}

func (r *repo) ContainsFile(filename string) bool {
	for _, entry := range r.rootBlock.Files {
		if entry.Name == filename {
			return true
		}
	}
	return false
}

func (r *repo) Put(filename string, opts *StorageOptions) error {

	finfo, err := os.Stat(filename)
	if err != nil {
		return err
	}

	if finfo.IsDir() {
		return errors.New("directories not supported")
	}

	if opts == nil {
		opts = r.config.DefaultStorageOpts(filename)
	}

	if r.ContainsFile(opts.FileName) {
		var name string
		for i := 1; ; i++ {
			name = fmt.Sprintf("%s (%d)", opts.FileName, i)
			if !r.ContainsFile(name) {
				break
			}
		}
		opts.FileName = name
	}

	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Determine how many blocks are in the file.
	blockInfos, err := getBlockInfo(file, r.config.BlockSize)
	if err != nil {
		return err
	}

	// Rewind the file.
	_, err = file.Seek(0, os.SEEK_SET)
	if err != nil {
		return err
	}

	pvdrinfo, err := r.listProviders()
	if err != nil {
		return err
	}

	var providers []core.Provider
	for _, pinfo := range pvdrinfo {
		pvdr, err := provider.Dial(pinfo.Addr)
		if err != nil {
			r.logger.Println("cannot dial", pinfo.Addr, "error:", err)
			continue
		}
		defer pvdr.Close()
		providers = append(providers, pvdr)
	}

	inode := core.INodeBlock{
		ID:      makeBlockId(r.config.UserId, opts.FileName),
		Name:    opts.FileName,
		OwnerID: r.config.UserId,
		Size:    finfo.Size(),
	}

	var contractInfos [][]contractInfo
	for _, binfo := range blockInfos {

		cinfos, err := r.negotiateContracts(binfo, providers)
		if err != nil {
			return fmt.Errorf("unable to negotiate storage contracts for block: %s", err)
		}
		contractInfos = append(contractInfos, cinfos)

		var contracts []*core.Contract
		for _, cinfo := range cinfos {
			contracts = append(contracts, cinfo.contract)
		}

		inode.Blocks = append(inode.Blocks, &core.BlockRef{
			ID:        binfo.ID,
			Contracts: contracts,
		})
	}

	// Negotiate contract for inode.
	cinfos, err := r.negotiateContracts(blockInfo{ID: inode.ID, Size: 1024 * 1024}, providers)
	if err != nil {
		return err
	}
	for _, cinfo := range cinfos {
		inode.Contracts = append(inode.Contracts, cinfo.contract)
	}

	// Upload file blocks
	for _, contracts := range contractInfos {
		block, err := readNextBlock(file, r.config.BlockSize)
		if err != nil {
			return err
		}

		for _, cinfo := range contracts {
			err := cinfo.provider.StoreBlock(cinfo.contract.BlockID, block)
			if err != nil {
				return err
			}
		}
	}

	// Upload inode block
	inodeBytes, err := marshalBlock(&inode)
	if err != nil {
		return err
	}

	for _, cinfo := range cinfos {
		err := cinfo.provider.StoreBlock(cinfo.contract.BlockID, inodeBytes)
		if err != nil {
			return err
		}
	}

	// Save inode to the repo cache
	err = saveBlock(path.Join(r.homedir, "user", inode.ID), &inode)
	if err != nil {
		return err
	}

	// Add record of the file to the user's root block
	r.rootBlock.Files = append(r.rootBlock.Files, &core.NamedBlockRef{ID: inode.ID, Name: inode.Name})
	err = saveBlock(path.Join(r.homedir, "user", r.rootBlock.ID), r.rootBlock)
	if err != nil {
		return err
	}

	return nil
}

func (r *repo) ListFiles() ([]string, error) {
	var res []string
	for _, blockInfo := range r.rootBlock.Files {
		res = append(res, blockInfo.Name)
	}
	return res, nil
}

func (r *repo) Get(filename string, out io.Writer) error {

	// Find locally cached inode for file.
	blockId := makeBlockId(r.config.UserId, filename)
	path := path.Join(r.homedir, "user", blockId)

	inode, err := loadINodeBlock(path)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.New("cannot find record of file " + filename)
		}
		return err
	}

	// Download file blocks.
	for _, blockRef := range inode.Blocks {
		data, err := r.downloadBlock(blockRef)
		if err != nil {
			return fmt.Errorf("cannot download file block. error: %s", err)
		}
		buf := bytes.NewBuffer(data)
		_, err = io.Copy(out, buf)
		if err != nil {
			return fmt.Errorf("cannot download file block. error: %s", err)
		}
	}
	return nil
}

func (r *repo) Sync() error {
	// TODO: Pull and merge updates to remote metadata.

	var providers []core.Provider
	if len(r.rootBlock.Contracts) == 0 {

		// Create storage contracts for root block.
		pinfos, err := r.listProviders()
		if err != nil {
			return err
		}

		var pvdrs []core.Provider
		for _, pinfo := range pinfos {
			pvdr, err := provider.Dial(pinfo.Addr)
			if err != nil {
				continue
			}
			defer pvdr.Close()
			pvdrs = append(pvdrs, pvdr)
		}

		binfo := blockInfo{
			ID:   r.rootBlock.ID,
			Size: 1024 * 1024,
		}

		cinfos, err := r.negotiateContracts(binfo, pvdrs)
		if err != nil {
			return err
		}

		for _, cinfo := range cinfos {
			r.rootBlock.Contracts = append(r.rootBlock.Contracts, cinfo.contract)
			providers = append(providers, cinfo.provider)
		}

	} else {
		for _, contract := range r.rootBlock.Contracts {
			pinfo, err := r.getProviderInfo(contract.ProviderID)
			if err != nil {
				continue
			}
			pvdr, err := provider.Dial(pinfo.Addr)
			if err != nil {
				continue
			}
			defer pvdr.Close()
			providers = append(providers, pvdr)
		}
	}

	if len(providers) == 0 {
		return errors.New("cannot connect to any metadata storage providers")
	}

	blockBytes, err := marshalBlock(&r.rootBlock)
	if err != nil {
		return err
	}

	nupdated := 0
	for _, pvdr := range providers {
		err := pvdr.StoreBlock(r.rootBlock.ID, blockBytes)
		if err != nil {
			continue
		}
		nupdated++
	}

	if nupdated < len(r.rootBlock.Contracts)/2 {
		return errors.New("unable to push metadata updates to enough providers")
	}

	return nil
}

func (r *repo) downloadBlock(ref *core.BlockRef) ([]byte, error) {
	for _, contract := range ref.Contracts {
		pinfo, err := r.getProviderInfo(contract.ProviderID)
		if err != nil {
			r.logger.Println("could not find provider info for", contract.ProviderID)
			continue
		}
		pvdr, err := provider.Dial(pinfo.Addr)
		if err != nil {
			r.logger.Println("could not dial provider", pinfo)
			continue
		}
		defer pvdr.Close()
		block, err := pvdr.GetBlock(ref.ID)
		if err != nil {
			r.logger.Println("could not download block", ref.ID, "error:", err)
			continue
		}
		return block, nil
	}
	return nil, errors.New("failed to download block")
}

func findHomeDir() (string, error) {
	skybinHome := os.Getenv("SKYBIN_HOME")

	if len(skybinHome) == 0 {
		user, err := user.Current()
		if err != nil {
			return "", fmt.Errorf("Cannot find skybin home directory: %s", err)
		}
		skybinHome = path.Join(user.HomeDir, ".skybin")
	}

	if err := checkRepo(skybinHome); err != nil {
		return "", err
	}

	return skybinHome, nil
}

func checkRepo(homedir string) error {
	_, err := os.Stat(homedir)
	if err != nil {
		return fmt.Errorf("Cannot find skybin home directory: %s", err)
	}
	return nil
}
