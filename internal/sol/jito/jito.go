package jito

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"example.com/m/internal/sol/utils"
	"example.com/m/pkg/logger"
	"github.com/gagliardetto/solana-go"
	jitorpc "github.com/jito-labs/jito-go-rpc"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type JitoSimResSuccess struct {
	Jsonrpc string `json:"jsonrpc"`
	Result  struct {
		Context struct {
			APIVersion string `json:"apiVersion"`
			Slot       int    `json:"slot"`
		} `json:"context"`
		Value struct {
			Summary            string `json:"summary"`
			TransactionResults []struct {
				Err                   any      `json:"err"`
				Logs                  []string `json:"logs"`
				PostExecutionAccounts any      `json:"postExecutionAccounts"`
				PreExecutionAccounts  any      `json:"preExecutionAccounts"`
				ReturnData            any      `json:"returnData"`
				UnitsConsumed         int      `json:"unitsConsumed"`
			} `json:"transactionResults"`
		} `json:"value"`
	} `json:"result"`
	ID string `json:"id"`
}

type JitoSimResFail struct {
	Jsonrpc string `json:"jsonrpc"`
	Result  struct {
		Context struct {
			APIVersion string `json:"apiVersion"`
			Slot       int    `json:"slot"`
		} `json:"context"`
		Value struct {
			Summary struct {
				Failed struct {
					Error struct {
						TransactionFailure []any `json:"TransactionFailure"`
					} `json:"error"`
					TxSignature string `json:"tx_signature"`
				} `json:"failed"`
			} `json:"summary"`
			TransactionResults []any `json:"transactionResults"`
		} `json:"value"`
	} `json:"result"`
	ID string `json:"id"`
}

func SimulateJitoBundle(ctx context.Context, transactions []solana.Transaction) error {

	endpoint := viper.GetString("HELIUS_URL")
	client := &http.Client{}

	encoded, err := utils.EncodeTransaction(transactions)
	if err != nil {
		return err
	}

	// Create the request payload properly
	requestPayload := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      "1",
		"method":  "simulateBundle",
		"params": []map[string]interface{}{
			{
				"encodedTransactions": encoded,
			},
		},
	}

	jsonData, err := json.Marshal(requestPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	var data = strings.NewReader(string(jsonData))
	req, err := http.NewRequest("POST", endpoint, data)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		bodyText, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(bodyText))
	}

	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	logger.Debug(nil, "jito sim res", zap.String("resp", string(bodyText)))
	var jitoResSuccess JitoSimResSuccess
	if err := json.Unmarshal(bodyText, &jitoResSuccess); err != nil {
		var jitoResFail JitoSimResFail
		if err := json.Unmarshal(bodyText, &jitoResFail); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
		errMsg, _ := json.Marshal(jitoResFail.Result.Value.Summary.Failed.Error.TransactionFailure)
		return fmt.Errorf("simulation failed: %s", errMsg)
	}

	return nil
}

func SendBundle(client *jitorpc.JitoJsonRpcClient, transactions []solana.Transaction) (string, error) {
	// Encode all transactions
	encodedTransactions, err := utils.EncodeTransaction(transactions)
	if err != nil {
		return "", err
	}

	// Create bundle request
	bundleRequest := [][]string{encodedTransactions}

	// Send the bundle
	bundleIdRaw, err := client.SendBundle(bundleRequest)
	if err != nil {
		return "", fmt.Errorf("failed to send bundle: %v", err)
	}

	// Decode response
	var bundleId string
	if err := json.Unmarshal(bundleIdRaw, &bundleId); err != nil {
		return "", fmt.Errorf("failed to unmarshal bundle ID: %v", err)
	}

	return bundleId, nil
}
