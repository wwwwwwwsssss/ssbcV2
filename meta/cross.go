package meta

type CrossTranParam struct {
	ContractName string   //智能合约名
	ContractFunc string   //智能合约函数
	ContractArgs []string //调用参数
}

type CrossTranProof struct {
	MerklePath  [][]byte //默克尔验证路径
	TransHash   []byte   //交易hash
	Height      int
	MerkleIndex []int64
}
