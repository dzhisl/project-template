package utils

import (
	"encoding/binary"

	"example.com/m/internal/constants"
	"github.com/gagliardetto/solana-go"
	associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"
	"github.com/gagliardetto/solana-go/programs/system"
)

func NewWrapSolInstruction(maker solana.PublicKey, amountToWrap uint64) []solana.Instruction {
	ixs := make([]solana.Instruction, 0)
	wsolAta := GetAta(constants.Wsol, maker)
	transferIx := system.NewTransferInstruction(
		amountToWrap,
		maker,
		wsolAta,
	)
	ixs = append(ixs, NewCreateATAIdempotentInstruction(maker, constants.Wsol), transferIx.Build(), newSyncNativeIx(wsolAta))

	return ixs
}

func NewCloseTokenAccountInstruction(maker, ata solana.PublicKey) *solana.GenericInstruction {
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
		maker, // Destination for remaining SOL
		true,  // Writable (receives SOL)
		false, // Not signer
	)

	ownerMeta := solana.NewAccountMeta(
		maker, // Owner/authority
		false, // Not writable
		true,  // Signer (must sign to authorize close)
	)

	ix := solana.NewInstruction(
		solana.TokenProgramID,
		[]*solana.AccountMeta{
			accountMeta,
			destinationMeta,
			ownerMeta,
		},
		[]byte{0x09}, // CloseAccount instruction discriminator
	)

	return ix
}

func newSyncNativeIx(ata solana.PublicKey) *solana.GenericInstruction {
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

func NewCreateATAIdempotentInstruction(wallet, mint solana.PublicKey) *solana.GenericInstruction {
	createATAInstruction := associatedtokenaccount.NewCreateInstruction(
		wallet,
		wallet,
		mint,
	).Build()

	return solana.NewInstruction(createATAInstruction.ProgramID(), createATAInstruction.Accounts(), []byte{1})
}

func NewPriorityFeeIx(totalFeeLamports uint64, computeUnits uint32) []solana.Instruction {
	ixs := make([]solana.Instruction, 0, 2)

	// 1. Set Compute Unit Limit
	computeUnitLimitIx := newSetComputeUnitLimitIx(computeUnits)
	ixs = append(ixs, computeUnitLimitIx)

	// 2. Calculate micro-lamports per compute unit
	// Formula: microLamportsPerCU = (totalFeeLamports * 1e6) / computeUnits
	microLamportsPerCU := (totalFeeLamports * 1_000_000) / uint64(computeUnits)

	// 3. Set Compute Unit Price (priority fee)
	computeUnitPriceIx := newSetComputeUnitPriceIx(microLamportsPerCU)
	ixs = append(ixs, computeUnitPriceIx)

	return ixs
}

func newSetComputeUnitLimitIx(computeUnits uint32) *solana.GenericInstruction {
	// SetComputeUnitLimit instruction data: [0x02] + 4 bytes for units (little endian)
	data := make([]byte, 5)
	data[0] = 0x02 // SetComputeUnitLimit discriminator
	binary.LittleEndian.PutUint32(data[1:], computeUnits)

	ix := solana.NewInstruction(
		solana.ComputeBudget,
		[]*solana.AccountMeta{}, // No accounts required
		data,
	)

	return ix
}

func newSetComputeUnitPriceIx(microLamports uint64) *solana.GenericInstruction {
	// SetComputeUnitPrice instruction data: [0x03] + 8 bytes for price (little endian)
	data := make([]byte, 9)
	data[0] = 0x03 // SetComputeUnitPrice discriminator
	binary.LittleEndian.PutUint64(data[1:], microLamports)

	ix := solana.NewInstruction(
		solana.ComputeBudget,
		[]*solana.AccountMeta{}, // No accounts required
		data,
	)

	return ix
}

func NewJitoTipIx(tipAmountLamports uint64, maker solana.PublicKey) *system.Instruction {
	nozomiAcc := solana.MustPublicKeyFromBase58("Cw8CFyM9FkoMi7K7Crf6HNQqf4uEMzpKw6QNghXLvLkY")
	transferix := system.NewTransferInstruction(tipAmountLamports, maker, nozomiAcc).Build()

	return transferix
}

func NewTransferIx(amount uint64, sender, receiver solana.PublicKey) solana.Instruction {
	return system.NewTransferInstruction(
		amount,
		sender,
		receiver,
	).Build()
}
