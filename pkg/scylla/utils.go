package scylla

import (
	"time"

	"github.com/gocql/gocql"
)

func CreateCluster(consistency gocql.Consistency, keyspace string, hosts ...string) *gocql.ClusterConfig {
	retryPolicy := &gocql.ExponentialBackoffRetryPolicy{
		Min:        time.Second,
		Max:        10 * time.Second,
		NumRetries: 5,
	}
	cluster := gocql.NewCluster(hosts...)
	cluster.Keyspace = keyspace
	cluster.Timeout = 5 * time.Second
	cluster.RetryPolicy = retryPolicy
	cluster.Consistency = consistency
	cluster.PoolConfig.HostSelectionPolicy = gocql.TokenAwareHostPolicy(gocql.RoundRobinHostPolicy())
	return cluster
}

// cluster := scylla.CreateCluster(gocql.Quorum, "upload", "scylla-node1", "scylla-node2", "scylla-node3")
// session, err := gocql.NewSession(*cluster)
// if err != nil {
// 	logger.Fatal("unable to connect to scylla", zap.Error(err))
// }

// upload.NewRepository(logger, session)
