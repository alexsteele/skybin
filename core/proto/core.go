package proto

//type EncryptionKey string
//
//type PeerInfo struct {
//	ID   string
//	Addr string
//}
//
//// Block contains the contents of a file.
//// A file may be stored as one or more blocks.
//type Block struct {
//	Data []byte
//}
//
//type BlockRef struct {
//	ID        string
//	Contracts []Contract
//}
//
//type INodeRef struct {
//	ID        string
//	Name      string
//	Contracts []Contract
//}
//
//type DirBlock struct {
//	ID           string
//	Contracts    []Contract
//	Name         string
//	LastModified time.Time
//	OwnerID      string
//	Files        []INodeRef
//}
//
//type INodeBlock struct {
//	ID             string
//	Contracts      []Contract
//	Name           string
//	OwnerID        string
//	AccessList     []string
//	EncryptionKeys []EncryptionKey
//	Blocks         []BlockRef
//	Size           int64
//}
//
//type Contract struct {
//	BlockID           string
//	BlockSize         int64
//	EndDate           time.Time
//	RenterID          string
//	ProviderID        string
//	RenterSignature   string
//	ProviderSignature string
//}

type Provider interface {
	Info() (*ProviderInfo, error)

	// Negotiate attempts to negotiate a storage contract. If the provider
	// agrees to the terms, the contract is  returned with the provider's
	// signature. Otherwise, the signature field is left empty, and the contract
	// is updated with the provider's requirements. If the provider is unwilling
	// to store the block, an error is returned.
	Negotiate(contract *Contract) (*Contract, error)

	// StoreBlock stores the given block with the provider.
	StoreBlock(id string, block []byte) error

	// GetBlock retrieves the given block from the provider.
	GetBlock(id string) (block []byte, err error)

	// TODO: Audit storage
}

//type ProviderInfo struct {
//	MaxBlockSize int
//	// Availability guarantees
//	// Accepted currencies
//	// Rates charged
//}
