package optimism

import "github.com/ethereum/hive/hivesim"

type Eth1Node struct {
	*hivesim.Client
	HTTPPort uint16
	WSPort   uint16
}

type DeployerNode struct {
	*hivesim.Client
	HTTPPort uint16
	WSPort   uint16
}

type L2Node struct {
	*hivesim.Client
	HTTPPort    uint16
	WSPort      uint16
	AuthrpcPort uint16
}

type OpNode struct {
	*hivesim.Client
	HTTPPort uint16
}

type L2OSNode struct {
	*hivesim.Client
}

type BSSNode struct {
	*hivesim.Client
}
