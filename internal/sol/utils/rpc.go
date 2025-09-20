package utils

import (
	"context"
	"fmt"
	"time"

	"example.com/m/pkg/logger"
	"github.com/gagliardetto/solana-go"
	"go.uber.org/zap"

	addresslookuptable "github.com/gagliardetto/solana-go/programs/address-lookup-table"
	"github.com/gagliardetto/solana-go/rpc"
)

var rpcClient *rpc.Client

func InitRpcUtils(client *rpc.Client) {
	rpcClient = client
}

func GetWalletBalance(wallet solana.PublicKey) (uint64, error) {
	balance, err := rpcClient.GetBalance(context.Background(), wallet, rpc.CommitmentProcessed)
	if err != nil {
		return 0, err
	}
	return balance.Value, nil
}

func GetBlockhash() (*solana.Hash, error) {
	hash, err := rpcClient.GetLatestBlockhash(context.Background(), rpc.CommitmentProcessed)
	if err != nil {
		return nil, err
	}
	return &hash.Value.Blockhash, nil
}

func GetLutAccount(account solana.PublicKey) (*addresslookuptable.AddressLookupTableState, error) {
	lut, err := addresslookuptable.GetAddressLookupTable(context.Background(), rpcClient, account)
	if err != nil {
		return nil, err
	}
	return lut, nil
}

func SendTransaction(tx *solana.Transaction) (string, error) {
	sig, err := rpcClient.SendTransactionWithOpts(context.Background(), tx, rpc.TransactionOpts{
		SkipPreflight:       false,
		PreflightCommitment: rpc.CommitmentProcessed,
	})
	if err != nil {
		return "", err
	}
	return sig.String(), nil
}

func ConfirmTransaction(signature solana.Signature) error {
	rpcClient := rpcClient

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	maxSupportedTransactionVersion := uint64(1)
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("transaction confirmation timeout after 30 seconds for signature: %s", signature.String())

		case <-ticker.C:
			logger.Debug(ctx, "polling transaction", zap.String("sig", signature.String()))
			txResult, err := rpcClient.GetTransaction(
				ctx,
				signature,
				&rpc.GetTransactionOpts{
					Commitment:                     rpc.CommitmentConfirmed,
					MaxSupportedTransactionVersion: &maxSupportedTransactionVersion,
				},
			)
			logger.Debug(ctx, "poll res", zap.Any("", txResult))
			if err != nil {
				logger.Error(ctx, "error in getting signature status", zap.Error(err))
				// If error is "not found", continue polling
				// Other errors might indicate network issues, continue polling as well
				continue
			}

			// Transaction found
			if txResult != nil {
				if txResult.Meta != nil && txResult.Meta.Err != nil {
					return fmt.Errorf("transaction failed with error: %v", txResult.Meta.Err)
				}
				return nil
			}
		}
	}
}
