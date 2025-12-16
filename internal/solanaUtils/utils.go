package solanaUtils

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"example.com/m/internal/constants"
	"example.com/m/pkg/cache"
	"example.com/m/pkg/config"
	"example.com/m/pkg/logger"
	"github.com/gagliardetto/solana-go"
	associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"
	computebudget "github.com/gagliardetto/solana-go/programs/compute-budget"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	"go.uber.org/zap"
)

const LamportsPerSol = 1000000000 // 1 SOL = 1 billion lamports

// sign a transaction with multiple wallets
func MultiSignTx(tx *solana.Transaction, wallets ...solana.PrivateKey) ([]solana.Signature, error) {
	return tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		for _, wallet := range wallets {
			if wallet.PublicKey().Equals(key) {
				return &wallet
			}
		}
		return nil
	})
}

func MultiPartialSignTx(tx *solana.Transaction, wallets ...solana.PrivateKey) ([]solana.Signature, error) {
	return tx.PartialSign(func(key solana.PublicKey) *solana.PrivateKey {
		for _, wallet := range wallets {
			if wallet.PublicKey().Equals(key) {
				return &wallet
			}
		}
		return nil
	})
}

func GetAta(mint, wallet, tokenProgram solana.PublicKey) (solana.PublicKey, error) {
	if tokenProgram.Equals(solana.Token2022ProgramID) {
		return getAta22(mint, wallet)
	}
	ata, _, err := solana.FindAssociatedTokenAddress(
		wallet,
		mint,
	)
	if err != nil {
		return solana.PublicKey{}, err
	}
	return ata, nil
}

func getAta22(mint, wallet solana.PublicKey) (solana.PublicKey, error) {
	// Derive ATA with Token-2022 Program ID (not legacy Token Program)
	ata, _, err := solana.FindProgramAddress(
		[][]byte{
			wallet.Bytes(),
			solana.Token2022ProgramID.Bytes(),
			mint.Bytes(),
		},
		solana.SPLAssociatedTokenAccountProgramID,
	)
	if err != nil {
		return solana.PublicKey{}, err
	}
	return ata, nil
}

// SolToLamports converts SOL to lamports
func SolToLamports(sol float64) uint64 {
	return uint64(sol * LamportsPerSol)
}

// LamportsToSol converts lamports to SOL
func LamportsToSol(lamports uint64) float64 {
	return float64(lamports) / LamportsPerSol
}

func GetLatestBlockhash(ctx context.Context) (*solana.Hash, error) {
	// 1. Пробуем достать из кэша
	v, err := cache.GetClient().Get("latest-blockhash")
	if err == nil && v != nil {
		if h, ok := v.(string); ok && h != "" {
			parsed := solana.MustHashFromBase58(h)
			return &parsed, nil
		}
	}

	// 2. Если нет — идём в RPC
	res, err := config.Get().Client().GetLatestBlockhash(ctx, rpc.CommitmentProcessed)
	if err != nil {
		return nil, err
	}

	hash := res.Value.Blockhash

	cache.GetClient().SetWithExpire("latest-blockhash", hash.String(), 10*time.Second)

	return &hash, nil
}

func GetBalance(ctx context.Context, wallet solana.PublicKey) (*rpc.GetBalanceResult, error) {
	return config.Get().Client().GetBalance(ctx, wallet, rpc.CommitmentConfirmed)
}

func GetTokenBalance(ctx context.Context, wallet, mint, tokenProgram solana.PublicKey) (uint8, uint64, float64, error) {
	ata, err := GetAta(mint, wallet, tokenProgram)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to get ata: %w", err)
	}
	return GetTokenAccountBalance(ctx, ata, tokenProgram)
}

func GetTokenAccountBalance(ctx context.Context, account, tokenProgram solana.PublicKey) (uint8, uint64, float64, error) {
	res, err := config.Get().Client().GetTokenAccountBalance(ctx, account, rpc.CommitmentProcessed)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to get token acc balance: %w", err)
	}
	tokenBalance, err := strconv.ParseFloat(res.Value.UiAmountString, 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to convert token amount from str to float64: %w", err)
	}
	tokenBalanceUint, err := strconv.Atoi(res.Value.Amount)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to convert token amount from str to float64: %w", err)
	}
	return res.Value.Decimals, uint64(tokenBalanceUint), tokenBalance, nil
}

func EncodeTransaction(txs []*solana.Transaction) ([]string, error) {
	encodedList := make([]string, 0, len(txs))
	for _, tx := range txs {
		serializedTx, err := tx.MarshalBinary()
		if err != nil {
			return nil, fmt.Errorf("failed to serialize transaction: %v", err)
		}
		encodedList = append(encodedList, base64.StdEncoding.EncodeToString(serializedTx))
	}
	return encodedList, nil
}

func NewMixerTransferIx(sender, receiver solana.PublicKey, lamportsToTransfer uint64) solana.Instruction {
	programID := solana.MustPublicKeyFromBase58("HcqjH2mXDts2b2yMfX112Lf1hRqLrkYUtX7L97dmNnAT")
	timestamp := uint64(time.Now().Unix())
	timestampBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(timestampBytes, timestamp)

	seeds := [][]byte{
		[]byte("temp"),
		sender.Bytes(),
		timestampBytes,
	}

	tempAccount, bump, err := solana.FindProgramAddress(seeds, programID)
	if err != nil {
		log.Fatalf("Failed to find PDA: %v", err)
	}

	// Generate 11 decoy accounts (random public keys)
	decoyAccounts := make([]solana.PublicKey, 11)
	for i := 0; i < 11; i++ {
		decoyAccounts[i] = solana.NewWallet().PublicKey()
	}
	instructionData := make([]byte, 18)
	instructionData[0] = 0 // discriminator
	binary.LittleEndian.PutUint64(instructionData[1:9], lamportsToTransfer)
	binary.LittleEndian.PutUint64(instructionData[9:17], timestamp)
	instructionData[17] = bump

	// Build account metas (15 accounts total)
	accounts := []*solana.AccountMeta{
		{PublicKey: sender, IsSigner: true, IsWritable: true},             // 0: Signer
		{PublicKey: tempAccount, IsSigner: false, IsWritable: true},       // 1: Temp account
		{PublicKey: receiver, IsSigner: false, IsWritable: true},          // 2: Receiver
		{PublicKey: system.ProgramID, IsSigner: false, IsWritable: false}, // 3: System program
	}

	// Add 11 decoy accounts (indices 4-14)
	for _, decoy := range decoyAccounts {
		accounts = append(accounts, &solana.AccountMeta{
			PublicKey:  decoy,
			IsSigner:   false,
			IsWritable: true,
		})
	}

	// Create the instruction
	instruction := solana.NewInstruction(
		programID,
		accounts,
		instructionData,
	)
	return instruction
}

func NewJitoTipIx(tipAmountLamports uint64, maker solana.PublicKey) *system.Instruction {
	return system.NewTransferInstruction(tipAmountLamports, maker, solana.MustPublicKeyFromBase58("Cw8CFyM9FkoMi7K7Crf6HNQqf4uEMzpKw6QNghXLvLkY")).Build()
}

func NewZeroSlotTipIx(tipAmountLamports uint64, maker solana.PublicKey) *system.Instruction {
	return system.NewTransferInstruction(tipAmountLamports, maker, solana.MustPublicKeyFromBase58("6rYLG55Q9RpsPGvqdPNJs4z5WTxJVatMB8zV3WJhs5EK")).Build()
}

func NewNozomiTipIx(tipAmountLamports uint64, maker solana.PublicKey) *system.Instruction {
	// min tip amount 0.001 SOL
	return system.NewTransferInstruction(tipAmountLamports, maker, solana.MustPublicKeyFromBase58("TEMPaMeCRFAS9EKF53Jd6KpHxgL47uWLcpFArU1Fanq")).Build()
}

func IsSigner(tx *solana.Transaction, acc solana.PublicKey) bool {
	list, err := tx.AccountMetaList()
	if err != nil {
		panic(err)
	}
	for _, i := range list {
		if i.IsSigner && i.PublicKey.String() == acc.String() {
			return true
		}
	}
	return false
}

func ConfirmTransaction(ctx context.Context, signature solana.Signature) error {
	return ConfirmTransactionWithDelay(ctx, signature, 1*time.Second)
}

func ConfirmTransactionWithDelay(ctx context.Context, signature solana.Signature, delay time.Duration) error {
	rpcClient := config.Get().Client()

	ticker := time.NewTicker(delay)
	defer ticker.Stop()
	maxSupportedTransactionVersion := uint64(1)
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("transaction confirmation timeout for signature: %s", signature.String())

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

			if err != nil {
				if !strings.Contains(err.Error(), "not found") {
					logger.Error(ctx, "error in getting signature status", zap.Error(err))
				}
				// If error is "not found", continue polling
				// Other errors might indicate network issues, continue polling as well
				continue
			}
			logger.Debug(ctx, "poll res", zap.Any("", txResult))

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

// CreatePriorityFeeInstruction creates an instruction for transaction priority fee
// priorityFee: total fee in lamports (1 SOL = 1e9 lamports)
// cu: compute units used by the transaction
func CreatePriorityFeeInstruction(priorityFee uint64, cu uint64) *computebudget.Instruction {
	// unitPrice = (priorityFee / cu) * 1e6
	// To avoid floating point operations, use integer math:
	unitPrice := (priorityFee * 1_000_000) / cu

	return computebudget.NewSetComputeUnitPriceInstruction(unitPrice).Build()
}

func TokensToFloat(tokenAmount, decimals uint64) float64 {
	return float64(tokenAmount) / math.Pow10(int(decimals))
}

func NewWrapSolInstruction(maker solana.PublicKey, amountToWrap uint64) ([]solana.Instruction, error) {
	ixs := make([]solana.Instruction, 0)
	wsolAta, err := GetAta(constants.Wsol, maker, solana.TokenProgramID)
	if err != nil {
		return nil, err
	}
	transferIx := system.NewTransferInstruction(
		amountToWrap,
		maker,
		wsolAta,
	)
	ixs = append(ixs, NewCreateATAIdempotentInstruction(maker, maker, constants.Wsol, solana.TokenProgramID), transferIx.Build(), NewSyncNativeIx(wsolAta))

	return ixs, nil
}

func NewSyncNativeIx(ata solana.PublicKey) *solana.GenericInstruction {
	meta := solana.NewAccountMeta(
		ata,
		true,
		false,
	)
	ix := solana.NewInstruction(
		solana.TokenProgramID,
		[]*solana.AccountMeta{
			meta,
		},
		[]byte{0x11},
	)
	return ix
}

func NewCreateATAIdempotentInstruction(payer, wallet, mint, tokenProgram solana.PublicKey) *solana.GenericInstruction {
	if tokenProgram.Equals(solana.Token2022ProgramID) {
		return newCreateATA22IdempotentInstruction(payer, wallet, mint)
	}
	createATAInstruction := associatedtokenaccount.NewCreateInstruction(
		payer,
		wallet,
		mint,
	).Build()

	return solana.NewInstruction(createATAInstruction.ProgramID(), createATAInstruction.Accounts(), []byte{1})
}

func newCreateATA22IdempotentInstruction(payer, wallet, mint solana.PublicKey) *solana.GenericInstruction {
	createATAInstruction := associatedtokenaccount.NewCreateInstruction(
		payer,
		wallet,
		mint,
	).Build()

	accounts := createATAInstruction.Accounts()

	// Derive ATA with Token-2022 Program ID (not legacy Token Program)
	ata, err := getAta22(mint, wallet)
	if err != nil {
		panic(err)
	}

	// Replace ATA address (index 1) and TokenProgram (index 5) with Token-2022
	accounts[1] = solana.NewAccountMeta(ata, true, false)
	accounts[5] = solana.NewAccountMeta(solana.Token2022ProgramID, false, false)

	return solana.NewInstruction(createATAInstruction.ProgramID(), accounts, []byte{1})
}

func NewCreateAccountWithSeedIx(maker solana.PublicKey, seed string) (solana.Instruction, error) {
	pubKey, seed, err := generatePubkeyWithSeed(maker, token.ProgramID, seed)
	if err != nil {
		return nil, fmt.Errorf("failed to generate pubkey with seed: %w", err)
	}

	// Create account
	createInst := system.NewCreateAccountWithSeedInstruction(
		maker,
		seed,
		SolToLamports(0.001),
		1,
		token.ProgramID,
		maker,
		pubKey,
		maker,
	)
	return createInst.Build(), nil
}

// generatePubkeyWithSeed creates a deterministic public key using a seed
func generatePubkeyWithSeed(maker solana.PublicKey, programId solana.PublicKey, seed string) (solana.PublicKey, string, error) {
	publicKey, err := solana.CreateWithSeed(maker, seed, programId)
	if err != nil {
		return solana.PublicKey{}, "", fmt.Errorf("failed to create with seed: %w", err)
	}

	return publicKey, seed, nil
}

func GetTokenDecimals(ctx context.Context, mint solana.PublicKey) (uint8, error) {
	res, err := config.Get().Client().GetTokenSupply(ctx, mint, rpc.CommitmentProcessed)
	if err != nil {
		return 0, err
	}
	return res.Value.Decimals, nil
}

// GetTokenTotalSupply returns the total supply of a token as uint64
func GetTokenTotalSupply(ctx context.Context, mint solana.PublicKey) (uint64, error) {
	res, err := config.Get().Client().GetTokenSupply(ctx, mint, rpc.CommitmentProcessed)
	if err != nil {
		return 0, err
	}
	amount, err := strconv.ParseUint(res.Value.Amount, 10, 64)
	if err != nil {
		return 0, err
	}
	return amount, nil
}

func GetTokenTopHolders(ctx context.Context, mint solana.PublicKey) ([]struct {
	PubKey    solana.PublicKey
	RawAmount string
}, error) {
	var m = make([]struct {
		PubKey    solana.PublicKey
		RawAmount string
	}, 0)
	res, err := config.Get().Client().GetTokenLargestAccounts(ctx, mint, rpc.CommitmentProcessed)
	if err != nil {
		return nil, err
	}
	for _, i := range res.Value {

		m = append(m, struct {
			PubKey    solana.PublicKey
			RawAmount string
		}{
			i.Address,
			i.Amount,
		})
	}
	return m, nil
}

// GetTokenAccountOwner returns the owner of a token account
func GetTokenAccountOwner(ctx context.Context, tokenAccount solana.PublicKey) (solana.PublicKey, error) {
	res, err := config.Get().Client().GetAccountInfoWithOpts(ctx, tokenAccount, &rpc.GetAccountInfoOpts{Commitment: rpc.CommitmentProcessed})
	if err != nil {
		return solana.PublicKey{}, err
	}

	if res == nil || res.Value == nil {
		return solana.PublicKey{}, fmt.Errorf("token account not found: %s", tokenAccount.String())
	}

	owner := res.Value.Owner
	return owner, nil
}

func NewCloseTokenAccountInstruction(maker, destination, ata, tokenProgram solana.PublicKey) *solana.GenericInstruction {
	// CloseAccount instruction requires:
	// 1. Account to close (the token account)
	// 2. Destination account (where remaining SOL goes - typically the owner)
	// 3. Owner of the token account (authority to close)

	accountMeta := solana.NewAccountMeta(
		ata,   // Account to close
		true,  // Writable (will be closed)
		false, // Not signer
	)

	destinationMeta := solana.NewAccountMeta(
		destination, // Destination for remaining SOL
		true,        // Writable (receives SOL)
		false,       // Not signer
	)

	ownerMeta := solana.NewAccountMeta(
		maker, // Owner/authority
		false, // Not writable
		true,  // Signer (must sign to authorize close)
	)

	ix := solana.NewInstruction(
		tokenProgram, // Use the correct token program (Token or Token-2022)
		[]*solana.AccountMeta{
			accountMeta,
			destinationMeta,
			ownerMeta,
		},
		[]byte{0x09}, // CloseAccount instruction discriminator
	)

	return ix
}

func RunBlockHashUpdater(ctx context.Context) {
	for {
		newHash, err := GetLatestBlockhash(ctx)
		if err != nil {
			logger.Error(ctx, "failed to get bh", zap.Error(err))
			time.Sleep(1 * time.Second)
			continue
		}

		// Store the blockhash in cache
		if err := cache.GetClient().Set("latest-blockhash", newHash.String()); err != nil {
			logger.Error(ctx, "failed to set blockhash in cache", zap.Error(err))
		}

		// logger.Debug(ctx, "bh updated", zap.String("new hash", newHash.String()))
		time.Sleep(1 * time.Second)
	}
}

// GeneratePrivateKeyFromSeed generates a deterministic Solana private key from a string seed
// seed: any string that will be used as entropy source
// Returns a Solana ed25519 private key derived from the seed using SHA-256
func GeneratePrivateKeyFromSeed(seed string) solana.PrivateKey {
	// Hash the seed string to get 32 bytes for ed25519
	hash := sha256.Sum256([]byte(seed))

	// Create ed25519 private key from the 32-byte seed
	privateKey := ed25519.NewKeyFromSeed(hash[:])

	return solana.PrivateKey(privateKey)
}

// NewTokenTransferIx creates a token transfer instruction that works with both Token Program and Token-2022
// mint: the token mint address
// sender: the sender's wallet public key
// receiver: the receiver's wallet public key
// amount: the amount of tokens to transfer (in raw token units, not adjusted for decimals)
// decimals: the token decimals (used for validation/logging purposes)
// tokenProgram: the token program ID (solana.TokenProgramID or solana.Token2022ProgramID)
func NewTokenTransferIx(mint, sender, receiver solana.PublicKey, amount uint64, decimals uint8, tokenProgram solana.PublicKey) (solana.Instruction, error) {
	// Get sender's associated token account
	senderAta, err := GetAta(mint, sender, tokenProgram)
	if err != nil {
		return nil, fmt.Errorf("failed to get sender ATA: %w", err)
	}

	// Get receiver's associated token account
	receiverAta, err := GetAta(mint, receiver, tokenProgram)
	if err != nil {
		return nil, fmt.Errorf("failed to get receiver ATA: %w", err)
	}

	// Create transfer instruction using the specified token program
	// Transfer instruction format (discriminator 0x03):
	// [0] = instruction discriminator (3 for Transfer)
	// [1..9] = amount (u64, little-endian)
	instructionData := make([]byte, 9)
	instructionData[0] = 0x03 // Transfer instruction
	binary.LittleEndian.PutUint64(instructionData[1:9], amount)

	accounts := []*solana.AccountMeta{
		{PublicKey: senderAta, IsSigner: false, IsWritable: true},   // Source account
		{PublicKey: receiverAta, IsSigner: false, IsWritable: true}, // Destination account
		{PublicKey: sender, IsSigner: true, IsWritable: false},      // Authority
	}

	return solana.NewInstruction(
		tokenProgram,
		accounts,
		instructionData,
	), nil
}

// NewBurnTokenIx creates a token burn instruction that works with both Token Program and Token-2022
// tokenAccount: the token account to burn from
// mint: the token mint address
// owner: the owner/authority of the token account
// amount: the amount of tokens to burn (in raw token units, not adjusted for decimals)
// tokenProgram: the token program ID (solana.TokenProgramID or solana.Token2022ProgramID)
func NewBurnTokenIx(tokenAccount, mint, owner solana.PublicKey, amount uint64, tokenProgram solana.PublicKey) solana.Instruction {
	// Create burn instruction using the specified token program
	// Burn instruction format (discriminator 0x08):
	// [0] = instruction discriminator (8 for Burn)
	// [1..9] = amount (u64, little-endian)
	instructionData := make([]byte, 9)
	instructionData[0] = 0x08 // Burn instruction
	binary.LittleEndian.PutUint64(instructionData[1:9], amount)

	accounts := []*solana.AccountMeta{
		{PublicKey: tokenAccount, IsSigner: false, IsWritable: true}, // Token account to burn from
		{PublicKey: mint, IsSigner: false, IsWritable: true},         // Mint (supply will decrease)
		{PublicKey: owner, IsSigner: true, IsWritable: false},        // Owner/authority
	}

	return solana.NewInstruction(
		tokenProgram,
		accounts,
		instructionData,
	)
}
