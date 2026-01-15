package node

import (
	"errors"
	"fmt"
	"io"
	"net"

	"time"

	raftlogger "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb/v2"
)

type store struct {
	m map[string]string
}

func (store) Apply(l *raft.Log) interface{} {
	return nil
}

func (store) Snapshot() (raft.FSMSnapshot, error) {
	return nil, nil
}

func (store) Restore(snapshot io.ReadCloser) error {
	return nil
}

func newStore() store {
	return store{}
}

type Node struct {
	raft *raft.Raft
	s    store
}

type Config struct {
	ID       string
	RaftAddr string
	JoinAddr string
}

func (c Config) validate() error {
	if c.ID == "" {
		return errors.New("must specify a node ID")
	}

	if c.RaftAddr == "" {
		return errors.New("must specify a raft bind address")
	}
	return nil
}

// New creates and starts up a node.
func New(c *Config) (*Node, error) {
	err := c.validate()
	if err != nil {
		return nil, err
	}

	raftConfig := raft.DefaultConfig()
	raftConfig.LocalID = raft.ServerID(c.ID)
	s := newStore()

	db, err := raftboltdb.NewBoltStore(fmt.Sprintf("./%s-db.log", c.ID))
	if err != nil {
		return nil, err
	}

	logger := raftlogger.Default()

	snapshotStore, err := raft.NewFileSnapshotStoreWithLogger("./", 10, logger)
	if err != nil {
		return nil, err
	}

	raftAddr, err := net.ResolveTCPAddr("tcp", c.RaftAddr)
	if err != nil {
		return nil, err
	}

	transport, err := raft.NewTCPTransportWithLogger(c.RaftAddr, raftAddr, 3, 10*time.Second, logger)
	if err != nil {
		return nil, err
	}

	r, err := raft.NewRaft(raftConfig, s, db, db, snapshotStore, transport)
	if err != nil {
		return nil, err
	}

	if c.JoinAddr == "" {
		r.BootstrapCluster(raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      raftConfig.LocalID,
					Address: transport.LocalAddr(),
				},
			},
		})
	}

	return &Node{
		raft: r,
		s:    s,
	}, nil
}

func (n *Node) Close() error {
	return nil
}

func (n *Node) Join(address string, nodeID string) error {
	// TODO: check that there isn't already a node with this id and address in the cluster.
	f := n.raft.AddVoter(raft.ServerID(nodeID), raft.ServerAddress(address), 0, 30*time.Second)
	if f.Error() != nil {
		return f.Error()
	}

	return nil
}
