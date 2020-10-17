package main

import (
	"bindexer/internal/utils"
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v4"
)

var (
	network                *string        = flag.String("network", "testnet", "Block Chain network to use")
	mode                   *string        = flag.String("mode", "update", "Daemon mode")
	osChan                 chan os.Signal = make(chan os.Signal, 1)
	blockHash, queryString string
	blockID                int64 = 0
	transactionID          int64 = 0
	transactionOutputID    int64 = 0
	nextBlock              string
	previousBlock          string
	service                *utils.Service
	rpcClient              *utils.RPCHttpClient
	connectionString       string
	tx                     pgx.Tx
	tables                 []string = []string{"blocks", "transaction_outputs", "transactions"}
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
	// Listen to this type of Interrupts
	signal.Notify(osChan, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGABRT, syscall.SIGBUS)
	// Build and Establish a database connection
	connectionString = fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		config.Database.Username,
		config.Database.Password,
		config.Database.Host,
		config.Database.Port,
		config.Database.Dbname)
	pgConn, err := pgx.Connect(context.Background(), connectionString)

	if err != nil {
		log.Fatal("Failed to connect to the database")
	}
	// Shutdown Indexer
	go func() {
		<-osChan
		if tx != nil {
			tx.Rollback(context.Background())
		}
		pgConn.Close(context.Background())
		log.Fatal("Indexer shutting down closing all resources")

	}()
	// Reindex if specified
	if *mode == "reindex" {
		reader := bufio.NewReader(os.Stdin)
		fmt.Println("Are you sure you want to wipeout the database (y/n) ? ")
		answer, _ := reader.ReadString('\n')
		if strings.ToLower(strings.TrimSpace(answer)) == "y" {
			tx, _ = pgConn.Begin(context.Background())

			for _, table := range tables {
				_, err = pgConn.Exec(context.Background(), fmt.Sprintf("TRUNCATE TABLE ONLY public.%s RESTART IDENTITY CASCADE;", table))
				if err != nil {
					log.Fatal("Failed to reindex blockchain : ", err)
				}
			}
			tx.Commit(context.Background())
		}
	}
	// Select the latest block from database if not start from gensis block
	queryString = `SELECT previous_block_hash,next_block_hash FROM blocks WHERE id = (SELECT max(id) FROM blocks) `
	err = pgConn.QueryRow(context.Background(), queryString).Scan(&previousBlock, &nextBlock)
	if err != nil {
		log.Println(err)
	}
	if strings.TrimSpace(previousBlock) == "" && strings.TrimSpace(nextBlock) == "" {
		gensisBlock, err := service.GetBlockHash(0)
		if err != nil {
			log.Fatal("Cant get the gensis block")
		}
		nextBlock = gensisBlock
	} else if strings.TrimSpace(nextBlock) == "" && strings.TrimSpace(previousBlock) != "" {
		nextBlock = previousBlock
	}
	// Loop through all the blocks
	for {
		timeStart := time.Now()
		block, err := service.GetBlock(nextBlock)
		// skips if does not exists
		if err != nil {
			log.Println(err)
			continue
		}
		// Wait for new blocks to come
		if strings.TrimSpace(block.NextBlockHash) == "" {
			continue
		}
		log.Printf("Block Hash: %s", block.Hash)
		log.Printf("Block Height: %d", block.Height)
		log.Printf("Confirmations: %d", block.Confirmations)
		log.Printf("Number of Transactions: %d", block.NumberOfTransactions)
		log.Println("Next Block Hash: ", block.NextBlockHash)

		// Begin transaction
		tx, err = pgConn.Begin(context.Background())
		if err != nil {
			log.Fatal("Error Begining transaction")
		}
		// Insert into block table
		queryString = `INSERT INTO blocks (hash, confirmations, size, weight, version, merkle_root, time,
					    median_time,height,difficulty,nonce,ntx,previous_block_hash, next_block_hash, indexed_at)
				        VALUES ($1, $2, $3, $4, $5,  $6, $7, $8,  $9, $10, $11, $12, $13, $14, NOW()) RETURNING id  ;
				    `
		err = pgConn.QueryRow(context.Background(), queryString,
			block.Hash,
			block.Confirmations,
			block.Size,
			block.Weight,
			block.Version,
			block.MerkleRoot,
			block.Time,
			block.MedianTime,
			block.Height,
			block.Difficulty,
			block.Nonce,
			block.NumberOfTransactions,
			block.PreviousBlockHash,
			block.NextBlockHash,
		).Scan(&blockID)

		if err != nil {
			tx.Rollback(context.Background())
			log.Fatal("Database Error", err)
		}
		// Skip gensis block transactions
		if block.Height > 0 {
			// Check if transactions exists
			for _, transaction := range block.Transactions {
				transaction, err := service.GetTransaction(transaction)
				if err != nil {
					tx.Rollback(context.Background())
					log.Fatal("Failed to Get Transaction ")
				}
				// Insert into transaction table
				queryString = `INSERT INTO transactions (txid, hash, version, hex, locktime, block_id, weight, block_time)  
						   VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id `
				err = pgConn.QueryRow(context.Background(), queryString,
					transaction.TransactionId,
					transaction.Hash,
					transaction.Version,
					transaction.Hex,
					transaction.LockTime,
					blockID,
					transaction.Weight,
					transaction.BlockTime,
				).Scan(&transactionID)

				if err != nil {
					tx.Rollback(context.Background())
					log.Fatal("Database Error : Transaction insert ", err)
				}
				// Insert into transaction output table

				for _, vout := range transaction.Vouts {
					queryString = `INSERT INTO transaction_outputs (output_transaction_id, value ,
							    vout_position ,script_pub_key_asm, script_pub_key_hex,script_pub_key_req_sigs, 
								script_pub_key_type, script_pub_key_addresses) 
								VALUES ($1, $2, $3, $4, $5, $6, $7, $8 )
   			                `
					_, err := pgConn.Exec(context.Background(), queryString,
						transactionID,
						vout.Value,
						vout.N,
						vout.ScriptPubKey.Asm,
						vout.ScriptPubKey.Hex,
						vout.ScriptPubKey.ReqSigs,
						vout.ScriptPubKey.Type,
						vout.ScriptPubKey.Addresses,
					)
					// Rollback if failed
					if err != nil {
						tx.Rollback(context.Background())
						log.Fatal("Database Error Transaction Output Insert", err)
					}
				}
				// Update Vins
				for _, vin := range transaction.Vins {
					queryString = `UPDATE transaction_outputs SET input_transaction_id = (SELECT id FROM transactions WHERE txid = $1),
						input_vout_position = $2 ,input_coinbase_id = $3  WHERE output_transaction_id = $4 `
					_, err = pgConn.Exec(context.Background(), queryString, vin.TransactionId, vin.Vout, vin.CoinBase, transactionID)

					// Rollback if failed
					if err != nil {
						tx.Rollback(context.Background())
						log.Fatal("Database Error Transaction Vin Output Update: ", err)

					}
				}
			}
		}
		// Commit Transaction
		err = tx.Commit(context.Background())
		if err != nil {
			log.Fatalf(" Database Transaction failed: %v", err)
		}
		timeEnd := time.Since(timeStart)
		log.Println("Time spent: ", timeEnd)
		log.Println("=================================================================================")
		// Advances to the Next Block
		nextBlock = block.NextBlockHash
	}
}
