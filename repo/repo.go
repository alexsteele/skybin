package repo

import (
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"skybin/core"
	pb "skybin/core/proto"
	"time"
)

type Config struct {
	UserId             string   `json:"userId"`
	DhtAddress         string   `json:"dhtAddress"`
	ApiAddress         string   `json:"apiAddress"`
	BootstrapAddresses []string `json:"bootstrapAddresses"`
	LogFolder          string   `json:"logFolder"`
}

type Repo interface {
	Store(filename string) error
	ListFiles() ([]string, error)
	Get(filename string, out io.Writer) error
	Config() *Config
	PeerStore() PeerStore
}

type repo struct {
	homedir   string
	config    *Config
	rootBlock *core.MetaBlock
}

func Load() (Repo, error) {
	homedir, err := findHomeDir()
	if err != nil {
		return nil, err
	}
	return LoadFrom(homedir)
}

func LoadFrom(homedir string) (Repo, error) {

	err := checkRepo(homedir)
	if err != nil {
		return nil, err
	}

	config, err := loadConfig(homedir)
	if err != nil {
		return nil, err
	}

	rootBlock, err := loadMetaBlock(homedir, makeMetaId(config.UserId, "/"))
	if err != nil {
		return nil, fmt.Errorf("Cannot load user's root block: %s", err)
	}

	return &repo{
		homedir:   homedir,
		config:    config,
		rootBlock: rootBlock,
	}, nil
}

func (r *repo) Store(filename string) error {

	providers, err := loadProviderList(r.homedir)
	if err != nil {
		return err
	}

	var client pb.PeerClient = nil
	var provider core.PeerInfo
	for _, peer := range providers {
		conn, err := grpc.Dial(peer.Addr, grpc.WithInsecure(), grpc.WithTimeout(3*time.Second))
		if err != nil {
			continue
		}
		defer conn.Close()
		client = pb.NewPeerClient(conn)
		provider = peer
		break

	}
	if client == nil {
		return fmt.Errorf("Cannot find storage provider")
	}

	blockData, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("Cannot not read file: %s", err)
	}

	blockId := hash(blockData)
	block := pb.Block{Data: blockData}
	request := pb.StoreBlockRequest{
		BlockId: blockId,
		Block:   &block,
	}

	_, err = client.StoreBlock(context.TODO(), &request)
	if err != nil {
		return fmt.Errorf("Cannot store block: %s", err)
	}

	contract := core.Contract{
		BlockID:    blockId,
		BlockSize:  int64(len(blockData)),
		RenterID:   r.Config().UserId,
		ProviderID: provider.ID,
	}

	blockInfo := core.BlockInfo{
		ID:        blockId,
		Contracts: []core.Contract{contract},
	}

	// Create an INode block for file
	metaBlock := core.MetaBlock{
		ID:           makeMetaId(r.Config().UserId, filename),
		Name:         filename,
		LastModified: time.Now(),
		OwnerID:      r.Config().UserId,
		Blocks:       []core.BlockInfo{blockInfo},
		Size:         int64(len(blockData)),
	}

	// Save it to the repo.
	err = saveMetaBlock(r.homedir, metaBlock.ID, metaBlock)
	if err != nil {
		return err
	}

	// Add record of the file to the user's root block
	r.rootBlock.Blocks = append(r.rootBlock.Blocks, core.BlockInfo{ID: metaBlock.ID, Name: filename})
	err = saveMetaBlock(r.homedir, r.rootBlock.ID, r.rootBlock)
	if err != nil {
		return err
	}

	return nil
}

func (r *repo) ListFiles() ([]string, error) {
	var res []string
	for _, blockInfo := range r.rootBlock.Blocks {
		res = append(res, blockInfo.Name)
	}
	return res, nil
}

func (r *repo) Get(filename string, out io.Writer) error {
	metaBlock, err := r.findMetaBlock(filename)
	if err != nil {
		return err
	}

	// TODO: In addition to being cached locally, provider addresses
	// will need to be searchable in the network.
	// We should probably also store a list of locations with each block.
	providers, err := loadProviderList(r.homedir)
	if err != nil {
		return err
	}

	providerMap := map[string]core.PeerInfo{}
	for _, provider := range providers {
		providerMap[provider.ID] = provider
	}

	for _, blockInfo := range metaBlock.Blocks {
		found := false
		for _, contract := range blockInfo.Contracts {
			providerInfo, exists := providerMap[contract.ProviderID]
			if !exists {
				continue
			}
			conn, err := grpc.Dial(providerInfo.Addr, grpc.WithInsecure())
			if err != nil {
				continue
			}
			defer conn.Close()
			client := pb.NewPeerClient(conn)
			req := pb.GetBlockRequest{BlockId: blockInfo.ID}
			resp, err := client.GetBlock(context.TODO(), &req)
			if err != nil {
				continue
			}
			// TODO: Check short write
			_, err = out.Write(resp.Block.Data)
			if err != nil {
				return err
			}
			found = true
			break

		}
		if !found {
			return errors.New("Unable to download file block")
		}
	}
	return nil
}

func (r *repo) findMetaBlock(filename string) (*core.MetaBlock, error) {
	for _, blockInfo := range r.rootBlock.Blocks {
		if blockInfo.Name == filename {
			return loadMetaBlock(r.homedir, blockInfo.ID)
		}
	}
	return nil, fmt.Errorf("Information for file %s not found", filename)
}

func (r *repo) Config() *Config {
	return r.config
}

func (r *repo) PeerStore() PeerStore {
	return &peerStore{
		peerdir: path.Join(r.homedir, "peer"),
	}
}

func makeMetaId(userId string, filename string) string {
	name := fmt.Sprintf("%s:%s", userId, filename)
	return hash([]byte(name))
}

func saveMetaBlock(homedir string, blockId string, block interface{}) error {
	data, err := json.MarshalIndent(block, "", "    ")
	if err != nil {
		return err
	}
	filename := path.Join(homedir, "user", blockId)
	return ioutil.WriteFile(filename, data, 0600)
}

func loadProviderList(homedir string) ([]core.PeerInfo, error) {
	path := path.Join(homedir, "providers.json")

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var providers []core.PeerInfo
	err = json.Unmarshal(bytes, &providers)
	return providers, err
}

func loadMetaBlock(homedir string, blockId string) (*core.MetaBlock, error) {
	path := path.Join(homedir, "user", blockId)

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block := &core.MetaBlock{}
	err = json.Unmarshal(bytes, block)
	if err != nil {
		return nil, err
	}

	return block, nil
}

func loadConfig(homedir string) (*Config, error) {
	configBytes, err := ioutil.ReadFile(path.Join(homedir, "config.json"))
	if err != nil {
		return nil, fmt.Errorf("Cannot read config: %s", err)
	}

	config := &Config{}
	err = json.Unmarshal(configBytes, config)
	if err != nil {
		return nil, fmt.Errorf("Cannot not parse config: %s", err)
	}
	return config, nil
}

func checkRepo(homedir string) error {
	_, err := os.Stat(homedir)
	if err != nil {
		return fmt.Errorf("Cannot find skybin home directory: %s", err)
	}
	return nil
}

func DefaultHomeDir() (string, error) {
	user, err := user.Current()
	if err != nil {
		return "", err
	}
	return path.Join(user.HomeDir, ".skybin"), nil
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
