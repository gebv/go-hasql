/*
   Copyright 2020 YANDEX LLC

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package hasql

import (
	"context"
	"database/sql"
	"fmt"
)

// Node of single cluster
type Node interface {
	fmt.Stringer

	Addr() string
	DB() *sql.DB
}

type sqlNode struct {
	addr string
	db   *sql.DB
}

var _ Node = &sqlNode{}

// NewNode constructs node from database/sql DB
func NewNode(addr string, db *sql.DB) Node {
	return &sqlNode{addr: addr, db: db}
}

// Addr returns node's address
func (n *sqlNode) Addr() string {
	return n.addr
}

// DB returns node's database/sql DB
func (n *sqlNode) DB() *sql.DB {
	return n.db
}

// String implements Stringer
func (n *sqlNode) String() string {
	return n.addr
}

// checkNode checks if the node is alive and whether it is primary or not
func checkNode(ctx context.Context, node Node, checker NodeChecker) (bool, error) {
	primary, err := checker(ctx, node.DB())
	if err != nil {
		return false, err
	}

	return primary, nil
}

// NodeStateCriteria for choosing a node
type NodeStateCriteria int

// Known node state criteria
const (
	Alive NodeStateCriteria = iota + 1
	Primary
	Standby
	PreferPrimary
	PreferStandby
)

// NodeChecker is a signature for functions that check if specific node is alive and is primary.
// Returns true for primary and false if not. If error is returned, node is considered dead.
// Check function can be used to perform a query returning single boolean value that signals
// if node is primary or not.
type NodeChecker func(ctx context.Context, db *sql.DB) (bool, error)

// NodePicker is a signature for functions that determine how to pick single node from set of nodes
type NodePicker func([]Node) Node
