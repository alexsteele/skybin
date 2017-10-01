package core

import (
	"time"
)

type EncryptionKey string

type PeerInfo struct {
	ID   string
	Addr string
}

// Block contains the contents of a file.
// A file may be stored as one or more blocks.
type Block struct {
	Data []byte
}

type BlockInfo struct {
	ID        string
	Name      string
	Contracts []Contract
}

// MetaBlock contains information about a file or folder
type MetaBlock struct {
	ID             string
	Contracts      []Contract // The contracts the metablock is stored under
	Name           string     // The name of the file or folder the metablock represents
	LastModified   time.Time
	OwnerID        string
	AccessList     []string // IDs of users allowed to access this file
	EncryptionKeys []EncryptionKey
	Blocks         []BlockInfo // Pointers to child blocks.
	Size           int64       // The size of the file this block represents
	Signature      string      // Digital signature of this block's SHA256 hash
}

type Contract struct {
	BlockID           string
	BlockSize         int64
	EndDate           time.Time
	RenterID          string
	ProviderID        string
	RenterSignature   string
	ProviderSignature string
}
