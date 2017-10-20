package remote

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	core "skybin/core/proto"
)

type RemoteProvider interface {
	core.Provider
	Close() error
}

func Dial(addr string) (RemoteProvider, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	pvdr := &remote{
		conn:   conn,
		client: core.NewProviderClient(conn),
	}
	return pvdr, nil
}

type remote struct {
	conn   *grpc.ClientConn
	client core.ProviderClient
}

func (p *remote) Info() (*core.ProviderInfo, error) {
	resp, err := p.client.Info(context.TODO(), &core.InfoRequest{})
	if err != nil {
		return nil, err
	}
	return resp.Info, nil
}

func (p *remote) Negotiate(contract *core.Contract) (*core.Contract, error) {
	resp, err := p.client.Negotiate(context.TODO(), &core.NegotiateRequest{
		Contract: contract,
	})
	if err != nil {
		return nil, err
	}
	return resp.Contract, nil
}

func (p *remote) StoreBlock(ID string, block []byte) error {
	_, err := p.client.StoreBlock(context.TODO(), &core.StoreBlockRequest{
		BlockId: ID,
		Block:   &core.Block{block},
	})
	return err
}

func (p *remote) GetBlock(ID string) (block []byte, err error) {
	resp, err := p.client.GetBlock(context.TODO(), &core.GetBlockRequest{
		BlockId: ID,
	})
	if err != nil {
		return nil, err
	}
	return resp.Block.Data, nil
}

func (p *remote) Close() error {
	return p.conn.Close()
}
