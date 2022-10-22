# sol-wallet-manager
Solana wallet manager. Made specfically for botters with a lot of burners.

# usage / setup
### config.json
rpc: input a valid, working RPC inside double quotes; preferably a non-rate limited RPC
sol: input a correct SOL address to collect SOL to
token: input a correct SOL address to collect tokens to
retries: amount of times to retry a transaction; default 3

---

### wallets.csv
format:
```
NAME,PRIVATE_KEY
wallet1_name,wallet1_private_key
wallet2_name,wallet2_private_key
```

# details
### Get Balance
shows the amount of SOL each wallet in `wallets.csv` has

### Collect SOL
sends all SOL from each wallet in `wallets.csv` to the SOL address in `sol` from `config.json`

please make sure the SOL address in the `sol` field is correct. the program will make you double check as well.

### Distribute SOL
checks and fills each select wallet up to a certain amount of sol.

ensure you have enough SOL in the sending wallet.

**if there are any issues with the conversion, open an issue and i will try to fix it. 

### Check Tokens
shows the amount of tokens (NFTs) each wallet in `wallets.csv` has

**note: this will also show empty token accounts in each wallet

### Transfer Tokens
sends tokens from each wallet in `wallets.csv` to the SOL address in `token` from `config.json`. this will also close the token account from the sending wallet in the same transaction (i.e. no empty token accounts are left behind)

please make sure the SOL address in the `token` field is correct. the program will make you double check as well.

### Close Empty Token Accounts
closes empty token accounts from each wallet in `wallets.csv`

### Show Keys
prints out the private keys from `wallets.csv`
