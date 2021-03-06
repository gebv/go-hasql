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

package hasql_test

import (
	"context"
	"database/sql"
	"time"

	"golang.yandex/hasql"
	"golang.yandex/hasql/checkers"
)

func ExampleNewCluster() {
	// cluster hosts
	hosts := []struct {
		Addr       string
		Connstring string
	}{
		{
			Addr:       "host1.yandex-team.ru",
			Connstring: "host=host1.yandex-team.ru",
		},
		{
			Addr:       "host2.yandex-team.ru",
			Connstring: "host=host2.yandex-team.ru",
		},
		{
			Addr:       "host3.yandex-team.ru",
			Connstring: "host=host3.yandex-team.ru",
		},
	}

	// Construct cluster nodes
	nodes := make([]hasql.Node, 0, len(hosts))
	for _, host := range hosts {
		// Create database pools for each node
		db, _ := sql.Open("pgx", host.Connstring)
		nodes = append(nodes, hasql.NewNode(host.Addr, db))
	}

	// Use options to fine-tune cluster behavior
	opts := []hasql.ClusterOption{
		hasql.WithUpdateInterval(2 * time.Second),        // set custom update interval
		hasql.WithNodePicker(hasql.PickNodeRoundRobin()), // set desired nodes selection algorithm
	}

	// Create cluster handler
	c, _ := hasql.NewCluster(nodes, checkers.PostgreSQL, opts...)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Wait for current primary
	node, err := c.WaitForPrimary(ctx)
	if err != nil {
		panic(err)
	}

	// Wait for any alive standby
	node, err = c.WaitForStandby(ctx)
	if err != nil {
		panic(err)
	}

	// Wait for any alive node
	node, err = c.WaitForAlive(ctx)
	if err != nil {
		panic(err)
	}
	// Wait for secondary node if possible, primary otherwise
	node, err = c.WaitForNode(ctx, hasql.PreferStandby)
	if err != nil {
		panic(err)
	}

	// Retrieve current primary
	node = c.Primary()
	if node == nil {
		panic("no primary")
	}
	// Retrieve any alive standby
	node = c.Standby()
	if node == nil {
		panic("no standby")
	}
	// Retrieve any alive node
	node = c.Alive()
	if node == nil {
		panic("everything is dead")
	}

	// Retrieve primary node if possible, secondary otherwise
	node = c.Node(hasql.PreferPrimary)
	if node == nil {
		panic("no primary nor secondary")
	}

	// Retrieve secondary node if possible, primary otherwise
	node = c.Node(hasql.PreferStandby)
	if node == nil {
		panic("no primary nor secondary")
	}

	// Do something on retrieved node
	if err = node.DB().PingContext(ctx); err != nil {
		panic(err)
	}
}
