package jito

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"example.com/m/internal/solanaUtils"
	router "example.com/m/pkg/httpRouter"
	"github.com/gagliardetto/solana-go"
	jitorpc "github.com/jito-labs/jito-go-rpc"
)

func SimulateJitoBundle(ctx context.Context, rpcUrl string, transactions []*solana.Transaction, skipSigVerify bool) error {
	client := &http.Client{}

	encoded, err := solanaUtils.EncodeTransaction(transactions)
	if err != nil {
		return err
	}

	var requestPayload map[string]interface{}

	if skipSigVerify {
		// Create preExecutionAccountsConfigs and postExecutionAccountsConfigs arrays
		// with null for each transaction (we don't need account state capture)
		accountConfigs := make([]interface{}, len(transactions))
		for i := range accountConfigs {
			accountConfigs[i] = nil
		}

		// Create the request payload with skipSigVerify and replaceRecentBlockhash
		requestPayload = map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      "1",
			"method":  "simulateBundle",
			"params": []interface{}{
				map[string]interface{}{
					"encodedTransactions": encoded,
				},
				map[string]interface{}{
					"preExecutionAccountsConfigs":  accountConfigs,
					"postExecutionAccountsConfigs": accountConfigs,
					"skipSigVerify":                true,
					"replaceRecentBlockhash":       true,
				},
			},
		}
	} else {
		// Use the old format without skipSigVerify options
		requestPayload = map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      "1",
			"method":  "simulateBundle",
			"params": []map[string]interface{}{
				{
					"encodedTransactions": encoded,
				},
			},
		}
	}

	status, resp, err := router.SendReqWithContext(ctx, client, "POST", rpcUrl, requestPayload, nil)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	if status != http.StatusOK {
		return fmt.Errorf("HTTP error %d: %s", status, string(resp))
	}

	// Parse the response to check for success
	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return fmt.Errorf("failed to parse response: %w - raw response: %s", err, string(resp))
	}

	// Check for JSON-RPC error
	if _, hasError := result["error"]; hasError {
		return fmt.Errorf("%s", string(resp))
	}

	// Check for success: only summary == "succeeded" indicates success
	if resultObj, ok := result["result"].(map[string]interface{}); ok {
		if valueObj, ok := resultObj["value"].(map[string]interface{}); ok {
			if summary, ok := valueObj["summary"].(string); ok && summary == "succeeded" {
				// This is the ONLY success indicator
				return nil
			}
		}
	}

	// Any other response is an error - return full response
	return fmt.Errorf("%s", string(resp))
}

func SendBundle(client *jitorpc.JitoJsonRpcClient, transactions []*solana.Transaction) (string, error) {
	// Encode all transactions
	encodedTransactions, err := solanaUtils.EncodeTransaction(transactions)
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
