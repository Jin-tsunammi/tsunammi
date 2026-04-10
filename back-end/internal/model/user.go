package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

const defaultUserRole = 1

type SendCode struct {
	Email string `json:"email"`
}

type ChangeEmailRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type SignUpWithEmail struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type SignInWithEmail struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type SignInJWTResp struct {
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
}

type SignInResp struct {
	User    *User          `json:"user"`
	JWTInfo *SignInJWTResp `json:"jwt_info"`
}

type IsUserExistReq struct {
	Email string `json:"email" validate:"required,email"`
}

type IsUserExistResp struct {
	Exist bool `json:"exist"`
}

type User struct {
	bun.BaseModel `bun:"table:users" swaggerignore:"true"`

	ID        uint64    `json:"id" bun:",pk,autoincrement" example:"1234567890"`
	Email     string    `json:"email" bun:"email,unique" example:"email@example.com"`
	PublicKey string    `json:"public_key" bun:"public_key,unique"`
	RoleId    uint64    `json:"role_id" bun:"role_id,notnull" example:"1"`
	CreatedAt time.Time `json:"created_at" bun:"created_at,notnull,nullzero,default:current_timestamp" example:"2025-10-01T03:12:20Z"`
	UpdatedAt time.Time `json:"updated_at" bun:"updated_at,notnull,nullzero,default:current_timestamp" example:"2025-10-01T03:12:20Z"`
}

func NewUser() *User {
	now := time.Now().UTC()

	return &User{
		RoleId:    defaultUserRole,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

type EmailVerificationCode struct {
	bun.BaseModel `bun:"table:codes" swaggerignore:"true"`

	ID        uint64    `json:"id" bun:",pk,autoincrement"`
	Email     string    `json:"email" bun:"email,notnull"`
	Code      string    `json:"code" bun:"code"`
	CreatedAt time.Time `json:"created_at" bun:"created_at,notnull,nullzero,default:current_timestamp"`
	ExpiresAt time.Time `json:"expires_at" bun:"expires_at,notnull"`
}

func NewVerificationCode(email string, code string) *EmailVerificationCode {
	now := time.Now().UTC()

	return &EmailVerificationCode{
		Email:     email,
		Code:      code,
		CreatedAt: now,
	}
}

type Session struct {
	bun.BaseModel `bun:"table:sessions" swaggerignore:"true"`

	ID        uuid.UUID `json:"id" bun:",pk,type:uuid,default:uuid_generate_v7()"`
	UserID    uint64    `json:"user_id" bun:"user_id,notnull"`
	Token     string    `json:"token" bun:"token,notnull"`
	CreatedAt time.Time `json:"created_at" bun:"created_at,notnull,nullzero,default:current_timestamp"`
	ExpiresAt time.Time `json:"expires_at" bun:"expires_at,notnull"`
}

func NewSession(sessionID uuid.UUID, userID uint64, token string, expiresAt time.Time) *Session {

	now := time.Now().UTC()

	return &Session{
		ID:        sessionID,
		UserID:    userID,
		Token:     token,
		CreatedAt: now,
		ExpiresAt: expiresAt,
	}
}

type UserAction string

const (
	ActionAddAPI          UserAction = "add_api"
	ActionDeleteAPI       UserAction = "delete_api"
	ActionWalletDeposit   UserAction = "wallet_deposit"
	ActionWalletImport    UserAction = "wallet_import"
	ActionWalletsBatchAdd UserAction = "wallets_batch_create"
)

type UserHistory struct {
	bun.BaseModel `bun:"table:user_actions" swaggerignore:"true"`

	ID        uint64     `json:"id" bun:",pk,autoincrement"`
	UserID    uint64     `json:"user_id" bun:"user_id,notnull"`
	Action    UserAction `json:"action" bun:"action,notnull"`
	Value     string     `json:"value" bun:"value,notnull"`
	CreatedAt time.Time  `json:"created_at" bun:"created_at,notnull,nullzero,default:current_timestamp"`
}

func NewUserHistory(userID uint64, action UserAction, value string) *UserHistory {
	now := time.Now().UTC()
	return &UserHistory{
		UserID:    userID,
		Action:    action,
		Value:     value,
		CreatedAt: now,
	}
}

type UserHistoryWithPaginationResponse struct {
	UserHistory []UserHistory `json:"user_history"`
	Page        int           `json:"page"`
	PageSize    int           `json:"page_size"`
	Total       int           `json:"total"`
}
