package confethcli

// EthChainID defines eth compatible chain id
// +genx:enum
type EthChainID uint64

const (
	ETH_CHAIN_ID_UNKNOWN      EthChainID = 0
	ETH_CHAIN_ID__ETH         EthChainID = 1         // Ethereum Mainnet
	ETH_CHAIN_ID__BSC         EthChainID = 56        // BNB Smart Chain
	ETH_CHAIN_ID__POLYGON     EthChainID = 137       // Polygon PoS
	ETH_CHAIN_ID__ARBITRUM    EthChainID = 42161     // Arbitrum One
	ETH_CHAIN_ID__OPTIMISM    EthChainID = 10        // OP Mainnet
	ETH_CHAIN_ID__AVALANCHE   EthChainID = 43114     // Avalanche C-Chain
	ETH_CHAIN_ID__BASE        EthChainID = 8453      // Base
	ETH_CHAIN_ID__TRON        EthChainID = 728126428 // Tron Mainnet
	ETH_CHAIN_ID__SEPOLIA     EthChainID = 11155111  // Sepolia
	ETH_CHAIN_ID__BSC_TESTNET EthChainID = 97        // BSC Testnet
)
