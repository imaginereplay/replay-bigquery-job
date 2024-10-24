package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Transaction struct {
	UserId                   string
	Day                      *big.Int
	Month                    *big.Int
	Year                     *big.Int
	AssetId                  string
	TotalDuration            *big.Int
	TotalRewardsConsumer     *big.Int
	TotalRewardsContentOwner *big.Int
}

func addToBlockchain(jobs []JobDataRow) error {
	client, err := ethclient.Dial(os.Getenv("RPC_URL"))
	if err != nil {
		log.Printf("Error connecting to Ethereum client: %v", err)
		return err
	}

	key := os.Getenv("DEPLOYER_PRIVATE_KEY")
	if strings.HasPrefix(key, "0x") {
		key = key[2:]
	}
	privateKey, err := crypto.HexToECDSA(key)
	if err != nil {
		log.Printf("Error converting private key: %v", err)
		return err
	}

	fromAddress := crypto.PubkeyToAddress(privateKey.PublicKey)

	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return err
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return err
	}

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return err
	}

	contractAddress := common.HexToAddress(os.Getenv("CONTRACT_ADDRESS"))
	parsedABI, err := abi.JSON(strings.NewReader(ABI))
	if err != nil {
		log.Printf("Error parsing ABI: %v", err)
		return err
	}

	contract := bind.NewBoundContract(contractAddress, parsedABI, client, client, client)

	transactions := make([]Transaction, len(jobs))
	for i, job := range jobs {
		assetID := ""
		if job.AssetID.Valid {
			assetID = job.AssetID.StringVal
		}

		totalRewardsConsumerWei := ToWei(job.TotalRewardsConsumer)
		totalRewardsContentOwnerWei := ToWei(job.TotalRewardsContentOwner)

		transactions[i] = Transaction{
			UserId:                   job.UserID,
			Day:                      big.NewInt(int64(job.CreatedAtDay.Day())),
			Month:                    big.NewInt(int64(job.CreatedAtDay.Month())),
			Year:                     big.NewInt(int64(job.CreatedAtDay.Year())),
			AssetId:                  assetID,
			TotalDuration:            big.NewInt(job.TotalDuration),
			TotalRewardsConsumer:     totalRewardsConsumerWei,
			TotalRewardsContentOwner: totalRewardsContentOwnerWei,
		}
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		return err
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)
	auth.GasPrice = gasPrice

	callData, err := parsedABI.Pack("batchInsertRecords", transactions)
	if err != nil {
		log.Printf("Error packing transaction data: %v", err)
		return err
	}

	msg := ethereum.CallMsg{
		From:     fromAddress,
		To:       &contractAddress,
		GasPrice: gasPrice,
		Value:    big.NewInt(0),
		Data:     callData,
	}

	gasLimit, err := client.EstimateGas(context.Background(), msg)
	if err != nil {
		log.Printf("Error estimating gas limit: %v", err)
		return err
	}

	auth.GasLimit = gasLimit

	tx, err := contract.Transact(auth, "batchInsertRecords", transactions)
	if err != nil {
		log.Printf("Error sending transaction: %v", err)
		return err
	}

	if tx == nil {
		log.Printf("Returned transaction is null")
		return fmt.Errorf("returned transaction is null")
	}

	receipt, err := waitForConfirmation(client, tx.Hash())
	if err != nil {
		log.Printf("Error waiting for transaction confirmation. Hash: %s, Error: %v", tx.Hash().Hex(), err)
		return err
	}

	if receipt.Status == 1 {
		log.Printf("Transaction successfully confirmed! Hash: %s", tx.Hash().Hex())
	} else {
		fmt.Printf("Transaction failed with status: %d, error: %v", receipt.Status, err)
		return fmt.Errorf("transaction failed with status: %d", receipt.Status)
	}

	return nil
}

func waitForConfirmation(client *ethclient.Client, txHash common.Hash) (*types.Receipt, error) {
	for {
		receipt, err := client.TransactionReceipt(context.Background(), txHash)
		if errors.Is(err, ethereum.NotFound) {
			time.Sleep(time.Second * 2)
			continue
		} else if err != nil {
			return nil, err
		}

		return receipt, nil
	}
}

func ToWei(ethValue float64) *big.Int {
	weiValue := new(big.Float).Mul(big.NewFloat(ethValue), big.NewFloat(1e18))
	weiBigInt := new(big.Int)
	weiValue.Int(weiBigInt)
	return weiBigInt
}
