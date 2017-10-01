package cmd

import (
	"bytes"
	"fmt"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path"
	pb "skybin/core/proto"
	skybinrepo "skybin/repo"
	"time"
)

var serverCmd = Cmd{
	Name:        "server",
	Description: "Run a background process to handle peer traffic",
	Usage:       "server",
	Run:         runServer,
}

type peerserver struct {
	store  skybinrepo.PeerStore
	logger *log.Logger
}

func (ps *peerserver) StoreBlock(ctxt context.Context, req *pb.StoreBlockRequest) (*pb.StoreBlockResponse, error) {
	ps.logger.Println("store block id", req.BlockId)
	err := ps.store.StoreBlock(req.BlockId, bytes.NewBuffer(req.Block.Data))
	return &pb.StoreBlockResponse{}, err
}

func (ps *peerserver) GetBlock(ctxt context.Context, req *pb.GetBlockRequest) (*pb.GetBlockResponse, error) {
	ps.logger.Println("get block id:", req.BlockId)
	br, err := ps.store.GetBlock(req.BlockId)
	if err != nil {
		return &pb.GetBlockResponse{}, err
	}
	defer br.Close()
	bytes, err := ioutil.ReadAll(br)
	if err != nil {
		return &pb.GetBlockResponse{}, err
	}
	return &pb.GetBlockResponse{
		Block: &pb.Block{Data: bytes},
	}, nil
}

func runServer(args []string) {

	repo, err := skybinrepo.Load()
	if err != nil {
		log.Fatal(err)
	}

	config := repo.Config()
	if err != nil {
		log.Fatal(err)
	}

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
	if len(config.LogFolder) > 0 {
		err := os.MkdirAll(config.LogFolder, 0666)
		if err != nil {
			log.Fatal("Cannot create log folder: ", err)
		}
		fname := fmt.Sprintf("skybin_log_%s.log", time.Now().Format("1-2-2006_15:04:05"))
		f, err := os.Create(path.Join(config.LogFolder, fname))
		if err != nil {
			log.Fatal("Cannot create log file: ", err)
		}
		defer f.Close()
		logger.SetOutput(f)
	}

	server := grpc.NewServer()
	peerServer := peerserver{
		store:  repo.PeerStore(),
		logger: logger,
	}
	pb.RegisterPeerServer(server, &peerServer)

	listener, err := net.Listen("tcp", config.ApiAddress)
	if err != nil {
		log.Fatalf("Cannot run API server at address %s: %s", config.ApiAddress, err)
	}
	logger.Println("Starting peer server at", config.ApiAddress)
	server.Serve(listener)

}
