package main

import (
	"bindexer/internal/engine"
	"bindexer/internal/utils"
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jackc/pgx/v4"
)

var (
	network          *string        = flag.String("network", "testnet", "Block Chain network to use")
	mode             *string        = flag.String("mode", "update", "Daemon mode")
	osChan           chan os.Signal = make(chan os.Signal, 1)
	service          *utils.Service
	rpcClient        *utils.RPCHttpClient
	pgConn           *pgx.Conn
	connectionString string
)

func main() {
	log.Println("Starting Bitcoin Indexer")
	flag.Parse()
	// Load configuration
	config, err := utils.LoadConfig("./config.json")
	if err != nil {
		log.Fatal("Fail to load config file")
	}
	// According to the network chosen
	if *network == "main" {
		rpcClient = rpcClient.New(config.Network.MainNet.Host, config.Network.MainNet.Username, config.Network.MainNet.Password, config.Network.MainNet.Port)
	} else {
		rpcClient = rpcClient.New(config.Network.TestNet.Host, config.Network.TestNet.Username, config.Network.TestNet.Password, config.Network.TestNet.Port)

	}
	// Service functions
	service = &utils.Service{
		Client: rpcClient,
	}

	// Build and Establish a database connection
	connectionString = fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		config.Database.Username,
		config.Database.Password,
		config.Database.Host,
		config.Database.Port,
		config.Database.Dbname)
	pgConn, err = pgx.Connect(context.Background(), connectionString)

	if err != nil {
		log.Fatal("Failed to connect to the database")
	}
	engine := &engine.Engine{Service: service, PgConn: pgConn}
	// Reindex if specified
	if *mode == "reindex" {
		reader := bufio.NewReader(os.Stdin)
		fmt.Println("Are you sure you want to wipeout the database (y/n) ? ")
		answer, _ := reader.ReadString('\n')
		if strings.ToLower(strings.TrimSpace(answer)) == "y" {
			engine.Reset()
		}
	}
	engine.Run()
}
