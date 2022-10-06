# Accumulate <> EVM Bridge Node
## Installation
1. Copy config into the server
```bash
mkdir ~/.accumulatebridge
curl -o ~/.accumulatebridge/config.yaml https://raw.githubusercontent.com/AccumulateNetwork/bridge/master/config.yaml.EXAMPLE
```

2. Fill configuration
```yaml
app:
# Node API port
  apiport: 8081
# Log Level (Debug 1, Info 2, Warn 3, Error 4)
  loglevel: 2
acme:
# Accumulate API endpoint, e.g. "https://testnet.accumulatenetwork.io/v2"
  node: ""
# Bridge ADI, usually "bridge.acme"
  bridgeadi: ""
# Accumulate ed25519 private key
  privatekey: ""
evm:
# EVM API endpoint (Infura/Quicknode, private node, etc.)
  node: ""
# EVM chainid (Ethereum mainnet 1, Goerli testnet 5, etc.)
  chainid: 1
# Gnosis safe smart contract address
  safeaddress: ""
# Accumulate bridge smart contract address
  bridgeaddress: ""
# EVM private key
  privatekey: ""
# (optional) Maximum gas fee (EIP-1559)
  maxgasfee: 30
# (optional) Maximum priority fee (EIP-1559)
  maxpriorityfee: 2
```

3. Change config owner
```bash
chown 1000:1000 ~/.accumulatebridge/config.yaml
```

4. Install using Docker (recommended)
```bash
docker run -d --name accumulatebridge -v ~/.accumulatebridge:/home/app/values registry.gitlab.com/accumulatenetwork/evm-bridge:main
```
