package cmd

import (
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"path"
	core "skybin/core/proto"
	provider "skybin/provider/local"
	skybinrepo "skybin/repo"
	"time"
)

var serverCmd = Cmd{
	Name:        "server",
	Description: "Run a provider server",
	Usage:       "server",
	Run:         runServer,
}

type server struct {
	provider core.Provider
	logger   *log.Logger
}

func (ps *server) Info(ctxt context.Context, req *core.InfoRequest) (*core.InfoResponse, error) {
	ps.logger.Println("get provider info")
	info, err := ps.provider.Info()
	if err != nil {
		return nil, err
	}
	return &core.InfoResponse{Info: info}, nil
}

func (ps *server) Negotiate(ctxt context.Context, req *core.NegotiateRequest) (*core.NegotiateResponse, error) {
	ps.logger.Println("create contract")
	contract, err := ps.provider.Negotiate(req.Contract)
	if err != nil {
		return nil, err
	}
	return &core.NegotiateResponse{Contract: contract}, nil
}

func (ps *server) StoreBlock(ctxt context.Context, req *core.StoreBlockRequest) (*core.StoreBlockResponse, error) {
	ps.logger.Println("store block id", req.BlockId)
	err := ps.provider.StoreBlock(req.BlockId, req.Block.Data)
	return &core.StoreBlockResponse{}, err
}

func (ps *server) GetBlock(ctxt context.Context, req *core.GetBlockRequest) (*core.GetBlockResponse, error) {
	ps.logger.Println("get block id:", req.BlockId)
	bytes, err := ps.provider.GetBlock(req.BlockId)
	if err != nil {
		return &core.GetBlockResponse{}, err
	}
	return &core.GetBlockResponse{
		Block: &core.Block{Data: bytes},
	}, nil
}

func runServer(args []string) {

	repo, err := skybinrepo.Open()
	if err != nil {
		log.Fatal(err)
	}

	rinfo := repo.Info()
	if err != nil {
		log.Fatal(err)
	}

	options := provider.Options{
		ProviderInfo: rinfo.Config.ProviderInfo,
		Dir:          path.Join(rinfo.HomeDir, "peer"),
	}

	provider, err := provider.New(options)
	if err != nil {
		log.Fatal(err)
	}

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
	if len(rinfo.Config.LogFolder) > 0 {
		err := os.MkdirAll(rinfo.Config.LogFolder, 0666)
		if err != nil {
			log.Fatal("cannot create log folder: ", err)
		}
		fname := fmt.Sprintf("provider_log_%s.log", time.Now().Format("1-2-2006_15:04:05"))
		f, err := os.Create(path.Join(rinfo.Config.LogFolder, fname))
		if err != nil {
			log.Fatal("cannot create log file: ", err)
		}
		defer f.Close()
		logger.SetOutput(f)
	}

	grpcServer := grpc.NewServer()
	server := server{
		provider: provider,
		logger:   logger,
	}
	core.RegisterProviderServer(grpcServer, &server)

	listener, err := net.Listen("tcp", rinfo.Config.ProviderAddress)
	if err != nil {
		log.Fatalf("cannot run API server at address %s: %s", listener.Addr(), err)
	}
	logger.Println("Starting provider server at", listener.Addr())
	log.Fatal(grpcServer.Serve(listener))
}
