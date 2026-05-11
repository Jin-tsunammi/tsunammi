package secret

import "fmt"

const VaultWalletMountPath = "secret/data/private_keys"
const VaultWalletMetadataPath = "secret/metadata/private_keys"
const VaultAccountMountPath = "secret/data/accounts"
const VaultAccountMetadataPath = "secret/metadata/accounts"

func CreateWalletSecretPath(userID uint64, publicKey string) string {
	return fmt.Sprintf("%s/%d/%s", VaultWalletMountPath, userID, publicKey)
}

func CreateWalletSecretMetadataPath(userID uint64, publicKey string) string {
	return fmt.Sprintf("%s/%d/%s", VaultWalletMetadataPath, userID, publicKey)
}

func CreateAccountSecretPath(userID, exchangeAccountID, accountID uint64) string {
	return fmt.Sprintf("%s/%d/%d/%d", VaultAccountMountPath, userID, exchangeAccountID, accountID)
}

func CreateAccountMetedataPath(userID, exchangeAccountID, accountID uint64) string {
	return fmt.Sprintf("%s/%d/%d/%d", VaultAccountMetadataPath, userID, exchangeAccountID, accountID)
}
