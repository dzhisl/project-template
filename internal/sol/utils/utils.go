package utils

import (
	"encoding/base64"
	"fmt"

	"github.com/gagliardetto/solana-go"
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

// GetLutAddressBySlot derives the PDA (Program Derived Address) for an Address Lookup Table
// based on the authority (owner) and recent slot
func GetLutAddressBySlot(slot uint64, owner solana.PublicKey) (solana.PublicKey, uint8, error) {
	// Convert slot to little-endian bytes
	slotBytes := make([]byte, 8)
	for i := 0; i < 8; i++ {
		slotBytes[i] = byte(slot >> (i * 8))
	}

	// Address Lookup Table Program ID
	addressLookupTableProgramID := solana.MustPublicKeyFromBase58("AddressLookupTab1e1111111111111111111111111")

	// Find PDA using authority and slot as seeds
	pda, bump, err := solana.FindProgramAddress(
		[][]byte{
			owner.Bytes(), // authority public key as seed
			slotBytes,     // slot as seed (little-endian)
		},
		addressLookupTableProgramID,
	)

	if err != nil {
		return solana.PublicKey{}, 0, err
	}

	return pda, bump, nil
}

func GetAta(mint, wallet solana.PublicKey) solana.PublicKey {
	ata, _, err := solana.FindAssociatedTokenAddress(
		wallet,
		mint,
	)
	if err != nil {
		panic(err)
	}
	return ata
}

// SolToLamports converts SOL to lamports
func SolToLamports(sol float64) uint64 {
	return uint64(sol * LamportsPerSol)
}

// LamportsToSol converts lamports to SOL
func LamportsToSol(lamports uint64) float64 {
	return float64(lamports) / LamportsPerSol
}

func EncodeTransaction(txs []solana.Transaction) ([]string, error) {
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
