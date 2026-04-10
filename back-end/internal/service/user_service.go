package service

import (
	"context"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"mm/internal/model"
	"mm/internal/storage/repository"
	"mm/pkg/apperrors"
	auth "mm/pkg/jwt"
	"mm/pkg/mtype"
	repo "mm/pkg/repository"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gagliardetto/solana-go"
)

const (
	referralTokenLength = 8
)

type UserService struct {
	UserRepository        *repository.UserRepository
	UserHistoryRepository *repository.UserHistoryRepository

	JWTAuth            auth.JWTAuthenticator
	TransactionManager *repo.TransactionManager
}

func NewUserService(
	userRepository *repository.UserRepository,
	userHistoryRepository *repository.UserHistoryRepository,
	jwtAuth auth.JWTAuthenticator,
	transactionManager *repo.TransactionManager,
) *UserService {
	return &UserService{
		UserRepository:        userRepository,
		UserHistoryRepository: userHistoryRepository,
		JWTAuth:               jwtAuth,
		TransactionManager:    transactionManager,
	}
}

func (s *UserService) CreateWithEmail(
	ctx context.Context,
	email mtype.Email,
) (*model.User, error) {

	user := model.NewUser()
	user.Email = email.String()

	res, err := s.UserRepository.CreateWithReturn(ctx, user)

	if repo.DuplicateKeyViolation(err) {
		return nil, apperrors.AlreadyExist("user with email already exists", err)
	}

	if err != nil {
		if repo.IsErrNoRows(err) {
			return nil, nil
		}
		return nil, apperrors.Internal("failed to create user", err)
	}

	return res, nil
}

func (s *UserService) CreateWithPublicAddress(
	ctx context.Context,
	address string,
) (*model.User, error) {
	user := model.NewUser()
	user.PublicKey = address

	user, err := s.UserRepository.CreateWithPublicAddress(ctx, user)

	if repo.DuplicateKeyViolation(err) {
		return nil, apperrors.AlreadyExist("user with email already exists", err)
	}

	if err != nil {
		return nil, apperrors.Internal("failed to create user", err)
	}

	return user, nil
}

func (s *UserService) GetByEmail(ctx context.Context, email mtype.Email) (*model.User, error) {
	user, err := s.UserRepository.FindByEmail(ctx, email)
	if err != nil {
		return nil, apperrors.NotFound("failed to find user by email", err)
	}

	if user == nil {
		return nil, apperrors.NotFound("user with email not found")
	}

	return user, nil
}

func (s *UserService) GetByPublicAddress(ctx context.Context, address string) (*model.User, error) {
	user, err := s.UserRepository.FindByPublicAddress(ctx, address)

	if err != nil {
		if repo.IsErrNoRows(err) {
			return nil, nil
		}
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetHistoryByUserID(ctx context.Context, userID uint64, page, pageSize int, from, to time.Time) (*model.UserHistoryWithPaginationResponse, error) {
	history, total, err := s.UserHistoryRepository.FetchAllByUserID(ctx, userID, page, pageSize, from, to)

	if err != nil {
		return nil, apperrors.Internal("failed to get history", err)
	}

	return &model.UserHistoryWithPaginationResponse{
		UserHistory: history,
		PageSize:    pageSize,
		Page:        page,
		Total:       total,
	}, nil
}

func (s *UserService) ChangeUserEmail(ctx context.Context, email mtype.Email, userID uint64) error {
	err := s.UserRepository.UpdateEmail(ctx, email, userID)

	if err != nil {
		return apperrors.Internal("failed to change user email", err)
	}

	return nil
}

func (s *UserService) DebugDeleteUser(ctx context.Context, userID uint64) error {
	return s.UserRepository.Delete(ctx, userID)
}

func VerifySecret(publicKey, secret string) bool {
	re := regexp.MustCompile(`[0-9]+`)
	digitsStrArray := re.FindAllString(publicKey, -1)

	sumOfDigits := 0
	for _, digitStr := range digitsStrArray {
		digit, err := strconv.Atoi(digitStr)
		if err != nil {
			return false
		}
		sumOfDigits += digit
	}

	totalSum := sumOfDigits - 1

	message := publicKey[:4] + publicKey[5:9] + publicKey[10:] + strconv.Itoa(totalSum)

	hash := sha512.Sum512([]byte(message))
	hashStr := hex.EncodeToString(hash[:])

	return strings.EqualFold(hashStr, secret)
}

func (s *UserService) CountAllUsers(
	ctx context.Context,
) (int, error) {
	count, err := s.UserRepository.CountUsers(ctx)
	if err != nil {
		return 0, apperrors.Internal("failed to count all users", err)
	}

	return count, nil
}
func (s *UserService) GetByID(
	ctx context.Context,
	id uint64,
) (*model.User, error) {
	user, err := s.UserRepository.GetByID(ctx, id)
	if err != nil {
		return nil, apperrors.Internal("failed to get user by id", err)
	}

	return user, nil
}

func (s *UserService) VerifyAddress(
	_ context.Context,
	req model.SolanaVerifyRequest,
) (bool, error) {
	publicKey, err := solana.PublicKeyFromBase58(req.PublicAddress)

	if err != nil {
		return false, apperrors.BadRequest("invalid solana address", err)
	}

	signatureBytes, err := base64.StdEncoding.DecodeString(req.SignedMessage)
	if err != nil {
		return false, apperrors.BadRequest("invalid signature format", err)
	}

	var signature solana.Signature
	copy(signature[:], signatureBytes)

	verified := signature.Verify(publicKey, []byte(req.PublicAddress))

	if !verified {
		return false, apperrors.BadRequest("invalid signature")
	}

	return true, nil
}
