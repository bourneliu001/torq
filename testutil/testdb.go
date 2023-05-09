package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"net/url"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/core"
	"github.com/lncapital/torq/internal/database"
	"github.com/lncapital/torq/internal/services_helpers"
	"github.com/lncapital/torq/internal/tags"
	prometheus2 "github.com/lncapital/torq/pkg/prometheus"
)

const superuserName = "postgres"
const testDbPort = 5433
const testDBPrefix = "torq_test_"
const TestPublicKey1 = "0326e692c411111111111111111111111111111111111111111111111111111111"
const TestPublicKey2 = "0326e692c422222222222222222222222222222222222222222222222222222222"
const TestFundingOutputIndex = 3
const TestFundingTransactionHash1 = "0101010101010101010101010101010101010101010101010101010101010101"
const TestChannelPoint1 = TestFundingTransactionHash1 + ":3"
const TestFundingTransactionHash2 = "0101010101010101010101010101010101010101010101010101010101010102"
const TestChannelPoint2 = TestFundingTransactionHash2 + ":3"
const TestFundingTransactionHash3 = "0101010101010101010101010101010101010101010101010101010101010103"
const TestChannelPoint3 = TestFundingTransactionHash3 + ":3"
const TestFundingTransactionHash4 = "0101010101010101010101010101010101010101010101010101010101010104"
const TestChannelPoint4 = TestFundingTransactionHash4 + ":3"
const TestFundingTransactionHash5_NOTINDB = "0101010101010101010101010101010101010101010101010101010101010105"
const TestChannelPoint5_NOTINDB = TestFundingTransactionHash5_NOTINDB + ":3"

func init() {
	// Set the seed for the random database name
	rand.Seed(time.Now().UnixNano())
}

// A Server represents a running PostgreSQL server.
type Server struct {
	baseURL string
	conn    *sql.DB
	dbNames []string
}

// InitTestDBConn creates a connection to the postgres user and creates the Server struct.
// This is used to create all other test databases and should be executed once at the top of a
// test file (in the Main function).
func InitTestDBConn() (*Server, error) {
	srv := &Server{
		baseURL: (&url.URL{
			Scheme: "postgres",
			Host:   fmt.Sprintf("localhost:%d", testDbPort),
			User:   url.UserPassword(superuserName, "password"),
			Path:   "/",
		}).String(),
	}

	var err error
	srv.conn, err = sql.Open("postgres", srv.baseURL+"?sslmode=disable")
	if err != nil {
		return nil, errors.Wrap(err, "SQL open connection")
	}

	//srv.conn.SetMaxOpenConns(1)

	return srv, nil
}

// Cleanup closes the connection to the connection to the postgres server used to create new test
// databases. This should only be used once for each test file.
func (srv *Server) Cleanup() error {

	killConnSql := `
		SELECT pg_terminate_backend(pid)
		FROM pg_stat_activity
		WHERE
			-- don't kill my own connection!
			pid <> pg_backend_pid()
			-- don't kill the connections to other databases
			AND datname LIKE '` + testDBPrefix + `%';`

	// Kill all connections before deleting the test_databases
	_, err := srv.conn.Exec(killConnSql)
	if err != nil {
		return errors.Wrapf(err, "srv.conn.Cleanup(%s)", killConnSql)
	}

	// Drop (delete) all test databases
	for _, name := range srv.dbNames {
		_, err := srv.conn.Exec("DROP DATABASE " + name + ";")
		if err != nil {
			return errors.Wrapf(err, "srv.conn.Cleanup(\"DROP DATABASE %s;\"", name)
		}
	}

	if srv.conn != nil {
		err = srv.conn.Close()
		if err != nil {
			return errors.Wrap(err, "Closing server database connection")
		}
	}

	return nil
}

// dbUrl creates the db url based on the db name.
func (srv *Server) dbUrl(dbName string) string {
	return srv.baseURL + dbName + "?sslmode=disable"
}

// randomString is used to generate a unique database names.
func randString(n int) string {
	// rune used as source for random database names
	var letters = []rune("abcdefghijklmnopqrstuvwxyz")

	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))] //nolint:gosec
	}
	return string(b)
}

// createDatabase creates a new database on the server and returns its
// data source name.
func (srv *Server) createDatabase() (string, error) {

	// Create a new random name for the test database with prefix.
	dbName := testDBPrefix + randString(16)

	// Create a new test database
	_, err := srv.conn.Exec("CREATE DATABASE " + dbName + ";")
	if err != nil {
		return "", errors.Wrapf(err, "srv.conn.ExecContext(ctx, \"CREATE DATABASE %s;\"", dbName)
	}
	log.Debug().Msgf("Created database: %v", dbName)

	// Store all database names so that they can be easily dropped (deleted)
	srv.dbNames = append(srv.dbNames, dbName)
	return srv.dbUrl(dbName), nil
}

// NewTestDatabase opens a connection to a freshly created database on the server.
func (srv *Server) NewTestDatabase() (*sqlx.DB, context.CancelFunc, error) {

	// Create the new test database based on the main server connection
	dns, err := srv.createDatabase()
	if err != nil {
		return nil, nil, errors.Wrap(err, "srv.createDatabase(ctx)")
	}

	// Connect to the new test database
	db, err := sqlx.Open("postgres", dns)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "sqlx.Open(\"postgres\", %s)", dns)
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	registry := prometheus.NewRegistry()
	services_helpers.SetMetrics(services_helpers.NewMetrics(registry))
	prometheus2.SetRegistry(registry)

	// Migrate the new test database
	err = database.MigrateUp(db)
	if err != nil {
		cancel()
		return nil, nil, errors.Wrap(err, "Database Migrate Up")
	}

	testNodeId1, err := addNode(db, TestPublicKey1, cancel)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Inserting default node for testing with publicKey: %v", TestPublicKey1)
	}

	testNodeId2, err := addNode(db, TestPublicKey2, cancel)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Inserting default node for testing with publicKey: %v", TestPublicKey2)
	}

	err = addNodeConnectionDetails(db, testNodeId1, cancel)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Inserting default node_connection_details for testing with nodeId: %v", testNodeId1)
	}

	err = addNodeConnectionDetails(db, testNodeId2, cancel)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Inserting default node_connection_details for testing with nodeId: %v", testNodeId2)
	}

	lndShortChannelId := uint64(1111)
	shortChannelId := core.ConvertLNDShortChannelID(lndShortChannelId)
	err = AddChannel(db, shortChannelId, lndShortChannelId, TestFundingTransactionHash1, TestFundingOutputIndex, testNodeId1, testNodeId2, cancel)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Inserting default channel for testing with shortChannelId: %v", shortChannelId)
	}

	lndShortChannelId = 2222
	shortChannelId = core.ConvertLNDShortChannelID(lndShortChannelId)
	err = AddChannel(db, shortChannelId, lndShortChannelId, TestFundingTransactionHash2, TestFundingOutputIndex, testNodeId1, testNodeId2, cancel)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Inserting default channel for testing with shortChannelId: %v", shortChannelId)
	}

	lndShortChannelId = 3333
	shortChannelId = core.ConvertLNDShortChannelID(lndShortChannelId)
	err = AddChannel(db, shortChannelId, lndShortChannelId, TestFundingTransactionHash3, TestFundingOutputIndex, testNodeId1, testNodeId2, cancel)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Inserting default channel for testing with shortChannelId: %v", shortChannelId)
	}

	lndShortChannelId = 4444
	shortChannelId = core.ConvertLNDShortChannelID(lndShortChannelId)
	err = AddChannel(db, shortChannelId, lndShortChannelId, TestFundingTransactionHash4, TestFundingOutputIndex, testNodeId1, testNodeId2, cancel)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Inserting default channel for testing with shortChannelId: %v", shortChannelId)
	}

	go cache.ChannelStatesCacheHandler(cache.ChannelStatesCacheChannel, ctx)
	go cache.SettingsCacheHandle(cache.SettingsCacheChannel, ctx)
	go cache.NodesCacheHandler(cache.NodesCacheChannel, ctx)
	go cache.NodeAliasesCacheHandler(cache.NodeAliasesCacheChannel, ctx)
	go cache.ChannelsCacheHandler(cache.ChannelsCacheChannel, ctx)
	go cache.TaggedCacheHandler(cache.TaggedCacheChannel, ctx)
	go cache.TriggersCacheHandler(cache.TriggersCacheChannel, ctx)
	go tags.TagsCacheHandler(tags.TagsCacheChannel, ctx)
	// TODO FIXME cyclic dependency so if you need this in tests then initialise it in the test
	//go automation.RebalanceCache(automation.ManagedRebalanceChannel, ctx)
	go cache.ServiceCacheHandler(cache.ServicesCacheChannel, ctx)

	cache.InitStates(true)
	_, cancelTorq := context.WithCancel(ctx)
	cache.InitRootService(cancelTorq)
	cache.SetActiveCoreServiceState(services_helpers.RootService)

	return db, cancel, nil
}

func addNode(db *sqlx.DB, testPublicKey string, cancel context.CancelFunc) (int, error) {
	var testNodeId int
	err := db.QueryRowx("INSERT INTO node (public_key, chain, network, created_on) VALUES ($1, $2, $3, $4) RETURNING node_id;",
		testPublicKey, core.Bitcoin, core.SigNet, time.Now().UTC()).Scan(&testNodeId)
	if err != nil {
		cancel()
		return 0, errors.Wrapf(err, "Inserting default node for testing with publicKey: %v", testPublicKey)
	}
	log.Debug().Msgf("Added test node with publicKey: %v nodeId: %v", testPublicKey, testNodeId)
	return testNodeId, nil
}

func addNodeConnectionDetails(db *sqlx.DB, testNodeId int, cancel context.CancelFunc) error {
	_, err := db.Exec(`INSERT INTO node_connection_details
			(node_id, name, implementation, status_id, ping_system, custom_settings, created_on, updated_on, node_start_date)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);`,
		testNodeId, fmt.Sprintf("Node_%v", testNodeId), core.LND, core.Active, 0, 0, time.Now().UTC(), time.Now().UTC(), nil)
	if err != nil {
		cancel()
		return errors.Wrapf(err, "Inserting default node_connection_details for testing with nodeId: %v", testNodeId)
	}
	log.Debug().Msgf("Added test active node connection details with nodeId: %v", testNodeId)
	return nil
}

func AddChannel(db *sqlx.DB, shortChannelId string, lndShortChannelId uint64,
	fundingTransactionHash string, fundingOutputIndex int,
	testNodeId1 int, testNodeId2 int, cancel context.CancelFunc) error {

	_, err := db.Exec(`INSERT INTO channel (
			  short_channel_id, funding_transaction_hash, funding_output_index,
			  closing_transaction_hash, closing_node_id,
			  lnd_short_channel_id, first_node_id, second_node_id, initiating_node_id, accepting_node_id, capacity, private,
			  status_id, created_on, updated_on, funding_block_height, funded_on, flags
			) values (
			  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
			);`,
		shortChannelId, fundingTransactionHash, fundingOutputIndex, nil, nil,
		lndShortChannelId, testNodeId1, testNodeId2, nil, nil, 1_000_000,
		false, core.Open, time.Now().UTC(), time.Now().UTC(), 10, time.Now().UTC(), 1)
	if err != nil {
		cancel()
		return errors.Wrapf(err, "Inserting default channel for testing with shortChannelId: %v", shortChannelId)
	}
	return nil
}
