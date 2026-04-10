package config

import (
	"time"

	"github.com/caarlos0/env/v11"
)

func NewConfig() *Config {
	config := env.Must[Config](env.ParseAs[Config]())

	return &config
}

const (
	EnvironmentProduction = "prod"
	EnvironmentStage      = "stage"
	EnvironmentDev        = "dev"
)

type Config struct {
	HTTP   HttpConfig
	DB     DBConfig
	App    AppConfig
	Crypto CryptoConfig
	Auth   AuthConfig
	Brevo  BrevoConfig
	Vault  VaultConfig
	Job    JobConfig
	Jito   JitoConfig
}

type HttpConfig struct {
	Host             string `env:"HTTP_HOST,required"`
	Port             string `env:"HTTP_PORT,required"`
	AllowOrigins     string `env:"ALLOW_ORIGINS,required"`
	AllowCredentials bool   `env:"ALLOW_CREDENTIALS,required"`
}

type CryptoConfig struct {
	WalletEncryptionKey  string `env:"WALLET_ENCRYPTION_KEY,required"`
	AccountEncryptionKey string `env:"ACCOUNT_ENCRYPTION_KEY,required"`
}

type AuthConfig struct {
	SecretAccessSignKey  string        `env:"SECRET_ACCESS_SIGN_KEY,required"`
	SecretRefreshSignKey string        `env:"SECRET_REFRESH_SIGN_KEY,required"`
	RefreshTokenTTL      time.Duration `env:"REFRESH_TOKEN_TTL,required"`
	AccessTokenTTL       time.Duration `env:"ACCESS_TOKEN_TTL,required"`
	VerificationCodeTTL  time.Duration `env:"VERIFICATION_CODE_TTL,required"`
}

type BrevoConfig struct {
	BrevoSecretKey string `env:"BREVO_SECRET_KEY,required"`
	BrevoEmail     string `env:"BREVO_EMAIL,required"`
	BrevoName      string `env:"BREVO_NAME,required"`
}

type VaultConfig struct {
	Address  string `env:"VAULT_ADDRESS,required"`
	RoleID   string `env:"VAULT_ROLE_ID,required"`
	SecretID string `env:"VAULT_SECRET_ID,required"`
}

type JobConfig struct {
	AccountsPendingCheckInterval      time.Duration `env:"ACCOUNTS_PENDING_CHECK_INTERVAL,required"`
	AccountsDeletedCheckInterval      time.Duration `env:"ACCOUNTS_DELETED_CHECK_INTERVAL,required"`
	WalletsPendingCreateCheckInterval time.Duration `env:"WALLETS_PENDING_CREATE_CHECK_INTERVAL,required"`
	WalletsPendingImportCheckInterval time.Duration `env:"WALLETS_PENDING_IMPORT_CHECK_INTERVAL,required"`
	TransactionPendingCheckInterval   time.Duration `env:"TRANSACTION_PENDING_CHECK_INTERVAL,required"`
	WithdrawLimitCheckInterval        time.Duration `env:"WITHDRAW_LIMIT_CHECK_INTERVAL" envDefault:"24h"`
}

type JitoConfig struct {
	RpcURLs           []string      `env:"JITO_RPC_URLS,required" env-separator:","`
	ProxyURLs         []string      `env:"JITO_PROXY_URLS" env-separator:","`
	CoolDown          time.Duration `env:"JITO_COOLDOWN" envDefault:"1080ms"`
	ScheduleTTL       time.Duration `env:"JITO_SCHEDULE_TTL" envDefault:"30m"`
	NetworkURL        string        `env:"JITO_NETWORK_URL" envDefault:"https://kobe.mainnet.jito.network/"`
	BundleURL         string        `env:"JITO_BUNDLE_URL" envDefault:"https://bundles.jito.wtf/"`
	SlotPadding       uint64        `env:"JITO_SLOT_PADDING" envDefault:"2"`
	BundleTimeout     time.Duration `env:"JITO_BUNDLE_TIMEOUT" envDefault:"15s"`
	ValidatorInterval time.Duration `env:"JITO_VALIDATOR_INTERVAL" envDefault:"90m"`
}

type AppConfig struct {
	Environment string `env:"ENVIRONMENT,required"`

	SolanaRPCURL string `env:"SOLANA_RPC_URL,required"`
	SolanaWSURL  string `env:"SOLANA_WS_URL,required"`

	KucoinBaseUrl string `env:"KUCOIN_BASE_URL" envDefault:"https://api.kucoin.com/"`

	JWTSecret string `env:"JWT_SECRET,required"`

	FirebaseFilePath string `env:"FIREBASE_FILE_PATH,required"`

	ExchangeRateCacheTTL time.Duration `env:"EXCHANGE_RATE_CACHE_TTL,required"`
	ExchangeRateURL      string        `env:"EXCHANGE_RATE_URL,required" envDefault:"https://lite-api.jup.ag/price/v3"`

	BlockhashInterval time.Duration `env:"BLOCKHASH_INTERVAL,required" envDefault:"1s"`

	RaydiumRPCURL string `env:"RAYDIUM_RPC_URL,required" envDefault:"https://api-v3.raydium.io"`

	SolscanApiKey string `env:"SOLSCAN_API_KEY,required"`
}
