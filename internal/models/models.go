package models

type Config struct {
	Database Credential `json:"database"`
	Network  Network    `json:"network"`
}
type Network struct {
	TestNet Credential `json:"testnet"`
	MainNet Credential `json:"main"`
}

type Credential struct {
	Host     string `json:"host"`
	Username string `json:"username"`
	Password string `json:"password"`
	Port     int64  `json:"port"`
	Dbname   string `json:"dbname,omitempty"`
}
type Block struct {
	Hash                 string   `json:"hash"`
	Confirmations        int64    `json:"confirmations"`
	Size                 int64    `json:"size"`
	Weight               int64    `json:"weight"`
	Height               int64    `json:"height"`
	Version              int32    `json:"version"`
	VersionHex           string   `json:"versionHex"`
	MerkleRoot           string   `json:"merkleroot"`
	Transactions         []string `json:"tx"`
	Time                 int64    `json:"time"`
	MedianTime           int64    `json:"mediantime"`
	Nonce                int64    `json:"nonce"`
	Bits                 string   `json:"bits"`
	Difficulty           float64  `json:"difficulty"`
	ChainWork            string   `json:"chainwork"`
	NumberOfTransactions int32    `json:"nTx"`
	PreviousBlockHash    string   `json:"previousblockhash,omitempty"`
	NextBlockHash        string   `json:"nextblockhash,omitempty"`
}

type Transaction struct {
	TransactionId string `json:"txid"`
	Hash          string `json:"hash"`
	Version       int64  `json:"version"`
	Size          int64  `json:"size"`
	Weight        int64  `json:"weight"`
	LockTime      int64  `json:"locktime"`
	Vins          []Vin  `json:"vin"`
	Vouts         []Vout `json:"vout"`
	Hex           string `json:"hex"`
	BlockHash     string `json:"blockhash"`
	Confirmations int64  `json:"confirmations"`
	Time          int64  `json:"time"`
	BlockTime     int64  `json:"blocktime"`
}

type Vin struct {
	TransactionId      string    `json:"txid"`
	Vout               uint16    `json:"vout"`
	CoinBase           string    `json:"coinbase,omitempty"`
	ScriptSig          ScriptSig `json:"scriptSig"`
	TransactionWitness []string  `json:"txinwitness"`
	Sequence           int64     `json:"sequence"`
}

type Vout struct {
	Value        float64      `json:"value"`
	N            uint16       `json:"n"`
	ScriptPubKey ScriptPubKey `json:"scriptPubKey"`
}

type ScriptPubKey struct {
	Asm       string   `json:"asm"`
	Hex       string   `json:"hex"`
	ReqSigs   uint16   `json:"reqSigs"`
	Type      string   `json:"type"`
	Addresses []string `json:"addresses"`
}

type ScriptSig struct {
	Asm string `json:"asm"`
	Hex string `json:"hex"`
}
