DROP TABLE IF EXISTS blocks;
-- CREATE block
CREATE TABLE blocks(
  id BIGSERIAL NOT NULL PRIMARY KEY,
  hash VARCHAR NOT NULL UNIQUE,
  confirmations BIGINT NOT NULL,
  size BIGINT NOT NULL,
  weight BIGINT NOT NULL,
  version BIGINT NOT NULL,
  merkle_root VARCHAR NOT NULL,
  time INTEGER NOT NULL,
  median_time BIGINT NOT NULL,
  height BIGINT NOT NULL,
  difficulty NUMERIC(15,2) NOT NULL,
  nonce BIGINT NOT NULL,
  ntx BIGINT NOT NULL,
  previous_block_hash VARCHAR,
  next_block_hash VARCHAR ,
  indexed_at TIMESTAMP NOT NULL
);
-- CREATE transactions
DROP TABLE IF EXISTS transactions;
CREATE TABLE transactions (
  id BIGSERIAL NOT NULL PRIMARY KEY,
  txid VARCHAR NOT NULL UNIQUE,
  hash VARCHAR NOT NULL UNIQUE,
  version INTEGER NOT NULL,
  hex VARCHAR NOT NULL,
  locktime BIGINT NOT NULL,
  block_id BIGINT NOT NULL REFERENCES blocks(id),
  weight BIGINT NOT NULL,
  block_time BIGINT NOT NULL
);

CREATE UNIQUE INDEX txid_idx ON transactions (txid);
CREATE UNIQUE INDEX hash_idx ON transactions (hash);

DROP TABLE IF EXISTS  transaction_outputs;
-- CREATE transaction outputs
CREATE TABLE transaction_outputs (
  id BIGSERIAL PRIMARY KEY NOT NULL, 
  output_transaction_id BIGINT REFERENCES transactions(id) NOT NULL,
  input_transaction_id BIGINT REFERENCES transactions(id) DEFAULT NULL,
  value NUMERIC(20,8) NOT NULL,
  input_vout_position INTEGER, 
  input_coinbase_id VARCHAR,
  vout_position INTEGER ,
  script_pub_key_asm VARCHAR ,
  script_pub_key_hex VARCHAR ,
  script_pub_key_req_sigs INTEGER ,
  script_pub_key_type VARCHAR,
  script_pub_key_addresses VARCHAR[]
);

