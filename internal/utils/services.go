package utils

import (
	"bindexer/internal/models"
	"bytes"
	"log"
	"strconv"
)

type Methods interface {
	GetBlock(gensisBlock string) (models.Block, error)
	GetTransaction(transactionHash string) (models.Transaction, error)
	GetBlockHash(height int64) (string, error)
	GetBlockCount() (int64, error)
}

type Service struct {
	Client *RPCHttpClient
}

// GetBlock when the blockHash is given
func (s *Service) GetBlock(blockHash string) (models.Block, error) {
	var req bytes.Buffer
	var block models.Block
	req.WriteString(`{"jsonrpc":"2.0",`)
	req.WriteString(`"method":`)
	req.WriteString(`"` + "getblock" + `",`)
	req.WriteString(`"params":[`)
	req.WriteString(`"` + blockHash + `"`)
	req.WriteString("]}")
	//log.Println(req.String())
	response, err := s.Client.DoRequest(req.String())
	if err != nil {
		return block, err
	}
	err = ConvertRpcResponseToType(response, &block)
	if err != nil {
		return block, err
	}
	return block, nil
}

// GetTransaction Details
// When a transactionHash is given
func (s *Service) GetTransaction(transactionHash string) (models.Transaction, error) {
	var req bytes.Buffer
	var transaction models.Transaction
	req.WriteString(`{"jsonrpc":"2.0",`)
	req.WriteString(`"method":`)
	req.WriteString(`"` + "getrawtransaction" + `",`)
	req.WriteString(`"params":[`)
	req.WriteString(`"` + transactionHash + `",`)
	req.WriteString("true")
	req.WriteString("]}")
	response, err := s.Client.DoRequest(req.String())
	if err != nil {
		return transaction, err
	}
	err = ConvertRpcResponseToType(response, &transaction)
	if err != nil {
		return transaction, err
	}
	return transaction, nil
}

// GetBlockHash Returns the block Hash when height is given
func (s *Service) GetBlockHash(height int64) (string, error) {
	var req bytes.Buffer
	var result string
	req.WriteString(`{"jsonrpc":"2.0",`)
	req.WriteString(`"method":`)
	req.WriteString(`"` + "getblockhash" + `",`)
	req.WriteString(`"params":[`)
	req.WriteString(strconv.FormatInt(height, 10))
	req.WriteString("]}")
	response, err := s.Client.DoRequest(req.String())
	if err != nil {
		return "", err
	}
	err = ConvertRpcResponseToType(response, &result)
	log.Println(result)
	if err != nil {
		return result, err
	}
	return result, nil
}

// GetBlockCount Returns the longest block in the chain
func (s *Service) GetBlockCount() (int64, error) {
	var req bytes.Buffer
	var result int64
	req.WriteString(`{"jsonrpc":"2.0",`)
	req.WriteString(`"method":`)
	req.WriteString(`"` + "getblockcount" + `",`)
	req.WriteString(`"params":[`)
	req.WriteString("]}")
	response, err := s.Client.DoRequest(req.String())
	if err != nil {
		return result, err
	}
	err = ConvertRpcResponseToType(response, &result)
	log.Println(result)
	if err != nil {
		return result, err
	}
	return result, nil
}
