package jito

// JitoErrorResponse represents an error response from the Jito API
type JitoErrorResponse struct {
	JsonRPC string `json:"jsonrpc"`
	Error   struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
	ID string `json:"id"`
}

// JitoFailedSimulation represents a failed simulation response
type JitoFailedSimulation struct {
	JsonRPC string `json:"jsonrpc"`
	Result  struct {
		Context struct {
			APIVersion string `json:"apiVersion"`
			Slot       int64  `json:"slot"`
		} `json:"context"`
		Value struct {
			Summary struct {
				Failed struct {
					Error struct {
						TransactionFailure []interface{} `json:"TransactionFailure"`
					} `json:"error"`
					TxSignature string `json:"tx_signature"`
				} `json:"failed"`
			} `json:"summary"`
			TransactionResults []interface{} `json:"transactionResults"`
		} `json:"value"`
	} `json:"result"`
	ID string `json:"id"`
}

// JitoSuccessSimulation represents a successful simulation response
type JitoSuccessSimulation struct {
	JsonRPC string `json:"jsonrpc"`
	Result  struct {
		Context struct {
			APIVersion string `json:"apiVersion"`
			Slot       int64  `json:"slot"`
		} `json:"context"`
		Value struct {
			Summary            string `json:"summary"`
			TransactionResults []struct {
				Err                   interface{} `json:"err"`
				Logs                  []string    `json:"logs"`
				PostExecutionAccounts interface{} `json:"postExecutionAccounts"`
				PreExecutionAccounts  interface{} `json:"preExecutionAccounts"`
				ReturnData            interface{} `json:"returnData"`
				UnitsConsumed         int         `json:"unitsConsumed"`
			} `json:"transactionResults"`
		} `json:"value"`
	} `json:"result"`
	ID string `json:"id"`
}
