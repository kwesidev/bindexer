package engine

import (
	"bindexer/internal/utils"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v4"
)

type Engine struct {
	Service *utils.Service
	PgConn  *pgx.Conn
}

func (e *Engine) Reset() {
	tables := []string{"blocks", "transaction_outputs", "transactions"}
	tx, _ := e.PgConn.Begin(context.Background())
	for _, table := range tables {
		_, err := e.PgConn.Exec(context.Background(), fmt.Sprintf("TRUNCATE TABLE ONLY public.%s RESTART IDENTITY CASCADE;", table))
		if err != nil {
			log.Fatal("Failed to reindex blockchain : ", err)
		}
	}
	tx.Commit(context.Background())
}

// Run  loops through the blocks and polls
func (e *Engine) Run() {

	var (
		blockID       int64 = 0
		transactionID int64 = 0
		nextBlock     string
		previousBlock string
		osChan        chan os.Signal = make(chan os.Signal, 1)
		tx            pgx.Tx
	)
	// Listen to this type of Interrupts
	signal.Notify(osChan, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGABRT, syscall.SIGBUS)

	// Shutdown Indexer if Interuppted
	go func() {
		<-osChan
		if tx != nil {
			tx.Rollback(context.Background())
		}
		e.PgConn.Close(context.Background())
		log.Fatal("Indexer shutting down closing all resources")

	}()
	queryString := `SELECT previous_block_hash,next_block_hash FROM blocks WHERE id = (SELECT max(id) FROM blocks) `
	err := e.PgConn.QueryRow(context.Background(), queryString).Scan(&previousBlock, &nextBlock)
	if err != nil {
		log.Println(err)
	}
	if strings.TrimSpace(previousBlock) == "" && strings.TrimSpace(nextBlock) == "" {
		gensisBlock, err := e.Service.GetBlockHash(0)
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
		block, err := e.Service.GetBlock(nextBlock)
		// skips if does not exists
		if err != nil {
			log.Println(err)
			continue
		}
		// Wait for new blocks to come
		if strings.TrimSpace(block.NextBlockHash) == "" {
			time.Sleep(20)
			continue
		}
		log.Printf("Block Hash: %s", block.Hash)
		log.Printf("Block Height: %d", block.Height)
		log.Printf("Confirmations: %d", block.Confirmations)
		log.Printf("Number of Transactions: %d", block.NumberOfTransactions)
		log.Println("Next Block Hash: ", block.NextBlockHash)

		// Begin transaction
		tx, err = e.PgConn.Begin(context.Background())
		if err != nil {
			log.Fatal("Error Begining transaction")
		}
		// Insert into block table
		queryString = `INSERT INTO blocks (hash, confirmations, size, weight, version, merkle_root, time,
					    median_time,height,difficulty,nonce,ntx,previous_block_hash, next_block_hash, indexed_at)
				        VALUES ($1, $2, $3, $4, $5,  $6, $7, $8,  $9, $10, $11, $12, $13, $14, NOW()) RETURNING id  ;
				    `
		err = e.PgConn.QueryRow(context.Background(), queryString,
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
				transaction, err := e.Service.GetTransaction(transaction)
				if err != nil {
					tx.Rollback(context.Background())
					log.Fatal("Failed to Get Transaction ")
				}
				// Insert into transaction table
				queryString = `INSERT INTO transactions (txid, hash, version, hex, locktime, block_id, weight, block_time)  
						   VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id `
				err = e.PgConn.QueryRow(context.Background(), queryString,
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
					_, err := e.PgConn.Exec(context.Background(), queryString,
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
					_, err = e.PgConn.Exec(context.Background(), queryString, vin.TransactionId, vin.Vout, vin.CoinBase, transactionID)

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
