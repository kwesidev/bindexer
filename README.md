Bitcoin Blockchain Indexer
==
A tool to index the bitcoin block chain and store it in PostgreSQL

Requirements
==
- PostgreSql >=9.3
- Bitcoind

Instructions
==
- Run the sqlscript file to create the tables
` psql -U postgres -d bitcoin < blockchain.sql `

- To run the Indexer 
Example  1
If you want to switch on  main network
`./bindexer --network=main `
If you want to run on testnet
Example 2
`./bindexer `

To run as  a service ,if using systemd then

`copy bindexer.service /usr/local/lib/systemd/system`