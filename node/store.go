package node

import (
	"errors"
	"fmt"
	"io"

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
	BindAddr string
}

func (c Config) validate() error {
	if c.ID == "" {
		return errors.New("must specify a node ID")
	}

	if c.BindAddr == "" {
		return errors.New("must specify a bind address")
	}

	return nil
}

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

	transport, err := raft.NewTCPTransportWithLogger(c.BindAddr, nil, 3, 10*time.Second, logger)
	if err != nil {
		return nil, err
	}

	r, err := raft.NewRaft(raftConfig, s, db, db, snapshotStore, transport)
	if err != nil {
		return nil, err
	}

	return &Node{
		raft: r,
		s:    s,
	}, nil
}
