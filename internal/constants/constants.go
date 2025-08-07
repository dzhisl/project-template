package constants

import (
	ag_solanago "github.com/gagliardetto/solana-go"
)

var (
	BonkProgramID           = ag_solanago.MustPublicKeyFromBase58("LanMV9sAd7wArD4vJFi2qDdfnVhFxYSUg6eADduJ3uj")
	Wsol                    = ag_solanago.MustPublicKeyFromBase58("So11111111111111111111111111111111111111112")
	RaydiumLpEventAuthority = ag_solanago.MustPublicKeyFromBase58("2DPAtwB8L12vrMRExbLuyGnC7n2J5LNoZQSejeQGpwkr")
	RaydiumLpAuthority      = ag_solanago.MustPublicKeyFromBase58("WLHv2UAZm6z4KyaaELi5pjdbJh6RESMva1Rnn8pJVVh")
)
