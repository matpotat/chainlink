package store

import (
	"context"
	"math/big"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/smartcontractkit/chainlink/core/store/assets"
	"github.com/smartcontractkit/chainlink/core/store/models"
	"github.com/smartcontractkit/chainlink/core/utils"
)

// EthClient holds the CallerSubscriber interface for the Ethereum blockchain.
type EthClient struct {
	CallerSubscriber
}

// CallerSubscriber implements the Call and EthSubscribe functions. Call performs
// a JSON-RPC call with the given arguments and EthSubscribe registers a subscription.
type CallerSubscriber interface {
	Call(result interface{}, method string, args ...interface{}) error
	EthSubscribe(context.Context, interface{}, ...interface{}) (models.EthSubscription, error)
}

// GetNonce returns the nonce (transaction count) for a given address.
func (eth *EthClient) GetNonce(address common.Address) (uint64, error) {
	result := ""
	err := eth.Call(&result, "eth_getTransactionCount", address.Hex(), "latest")
	if err != nil {
		return 0, err
	}
	return utils.HexToUint64(result)
}

// GetWeiBalance returns the balance of the given address in Wei.
func (eth *EthClient) GetWeiBalance(address common.Address) (*big.Int, error) {
	result := ""
	numWeiBigInt := new(big.Int)
	err := eth.Call(&result, "eth_getBalance", address.Hex(), "latest")
	if err != nil {
		return numWeiBigInt, err
	}
	numWeiBigInt.SetString(result, 0)
	return numWeiBigInt, nil
}

// GetEthBalance returns the balance of the given addresses in Ether.
func (eth *EthClient) GetEthBalance(address common.Address) (*assets.Eth, error) {
	balance, err := eth.GetWeiBalance(address)
	if err != nil {
		return assets.NewEth(0), err
	}
	return (*assets.Eth)(balance), nil
}

// callArgs represents the data used to call the balance method of an ERC
// contract. "To" is the address of the ERC contract. "Data" is the message sent
// to the contract.
type callArgs struct {
	To   common.Address `json:"to"`
	Data hexutil.Bytes  `json:"data"`
}

// ExtractERC20BalanceTargetAddress returns the address whose balance is being
// queried by the message in the given call to an ERC20 contract, which is
// interpreted as a callArgs.
func ExtractERC20BalanceTargetAddress(args interface{}) (common.Address, bool) {
	call, ok := (args).(callArgs)
	if !ok {
		return common.Address{}, false
	}
	message := call.Data
	return common.BytesToAddress(([]byte)(message)[len(message)-20:]), true
}

// GetERC20Balance returns the balance of the given address for the token contract address.
func (eth *EthClient) GetERC20Balance(address common.Address, contractAddress common.Address) (*big.Int, error) {
	result := ""
	numLinkBigInt := new(big.Int)
	functionSelector := models.HexToFunctionSelector("0x70a08231") // balanceOf(address)
	data := utils.ConcatBytes(functionSelector.Bytes(), common.LeftPadBytes(address.Bytes(), utils.EVMWordByteLen))
	args := callArgs{
		To:   contractAddress,
		Data: data,
	}
	err := eth.Call(&result, "eth_call", args, "latest")
	if err != nil {
		return numLinkBigInt, err
	}
	numLinkBigInt.SetString(result, 0)
	return numLinkBigInt, nil
}

// SendRawTx sends a signed transaction to the transaction pool.
func (eth *EthClient) SendRawTx(hex string) (common.Hash, error) {
	result := common.Hash{}
	err := eth.Call(&result, "eth_sendRawTransaction", hex)
	return result, err
}

// GetTxReceipt returns the transaction receipt for the given transaction hash.
func (eth *EthClient) GetTxReceipt(hash common.Hash) (*models.TxReceipt, error) {
	receipt := models.TxReceipt{}
	err := eth.Call(&receipt, "eth_getTransactionReceipt", hash.String())
	return &receipt, err
}

// GetBlockNumber returns the block number of the chain head.
func (eth *EthClient) GetBlockNumber() (uint64, error) {
	result := ""
	if err := eth.Call(&result, "eth_blockNumber"); err != nil {
		return 0, err
	}
	return utils.HexToUint64(result)
}

// GetBlockByNumber returns the block for the passed hex, or "latest", "earliest", "pending".
func (eth *EthClient) GetBlockByNumber(hex string) (models.BlockHeader, error) {
	var header models.BlockHeader
	err := eth.Call(&header, "eth_getBlockByNumber", hex, false)
	return header, err
}

// GetLogs returns all logs that respect the passed filter query.
func (eth *EthClient) GetLogs(q ethereum.FilterQuery) ([]models.Log, error) {
	var results []models.Log
	err := eth.Call(&results, "eth_getLogs", utils.ToFilterArg(q))
	return results, err
}

// GetChainID returns the ethereum ChainID.
func (eth *EthClient) GetChainID() (uint64, error) {
	var intermediary models.Big
	err := eth.Call(&intermediary, "eth_chainId")
	return intermediary.ToInt().Uint64(), err
}

// SubscribeToLogs registers a subscription for push notifications of logs
// from a given address.
func (eth *EthClient) SubscribeToLogs(
	channel chan<- models.Log,
	q ethereum.FilterQuery,
) (models.EthSubscription, error) {
	// https://github.com/ethereum/go-ethereum/blob/762f3a48a00da02fe58063cb6ce8dc2d08821f15/ethclient/ethclient.go#L359
	ctx := context.Background()
	sub, err := eth.EthSubscribe(ctx, channel, "logs", utils.ToFilterArg(q))
	return sub, err
}

// SubscribeToNewHeads registers a subscription for push notifications of new blocks.
func (eth *EthClient) SubscribeToNewHeads(
	channel chan<- models.BlockHeader,
) (models.EthSubscription, error) {
	ctx := context.Background()
	sub, err := eth.EthSubscribe(ctx, channel, "newHeads")
	return sub, err
}
