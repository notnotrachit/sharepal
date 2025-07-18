package models

import (
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
)

var passwordRule = []validation.Rule{
	validation.Required,
	validation.Length(8, 32),
	validation.Match(regexp.MustCompile("^\\S+$")).Error("cannot contain whitespaces"),
}

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (a RegisterRequest) Validate() error {
	return validation.ValidateStruct(&a,
		validation.Field(&a.Name, validation.Required, validation.Length(3, 64)),
		validation.Field(&a.Email, validation.Required, is.Email),
		validation.Field(&a.Password, passwordRule...),
	)
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (a LoginRequest) Validate() error {
	return validation.ValidateStruct(&a,
		validation.Field(&a.Email, validation.Required, is.Email),
		validation.Field(&a.Password, passwordRule...),
	)
}

type RefreshRequest struct {
	Token string `json:"token"`
}

func (a RefreshRequest) Validate() error {
	return validation.ValidateStruct(&a,
		validation.Field(
			&a.Token,
			validation.Required,
			validation.Match(regexp.MustCompile("^\\S+$")).Error("cannot contain whitespaces"),
		),
	)
}

type NoteRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (a NoteRequest) Validate() error {
	return validation.ValidateStruct(&a,
		validation.Field(&a.Title, validation.Required),
		validation.Field(&a.Content, validation.Required),
	)
}

// Group related requests
type CreateGroupRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Currency    string   `json:"currency"`
	MemberIDs   []string `json:"member_ids,omitempty"`
}

func (r CreateGroupRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Name, validation.Required, validation.Length(1, 100)),
		validation.Field(&r.Currency, validation.Required, validation.Length(3, 3)),
	)
}

type UpdateGroupRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Currency    string `json:"currency,omitempty"`
}

func (r UpdateGroupRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Name, validation.Length(1, 100)),
		validation.Field(&r.Currency, validation.Length(3, 3)),
	)
}

type AddMemberToGroupRequest struct {
	UserID string `json:"user_id"`
}

func (r AddMemberToGroupRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.UserID, validation.Required),
	)
}

// Friendship related requests
type SendFriendRequestRequest struct {
	Email string `json:"email"`
}

func (r SendFriendRequestRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Email, validation.Required, is.Email),
	)
}

type RespondFriendRequestRequest struct {
	Accept bool `json:"accept"`
}

func (r RespondFriendRequestRequest) Validate() error {
	return validation.ValidateStruct(&r)
}

// Transaction related requests
type TransactionPayerRequest struct {
	UserID string  `json:"user_id"`
	Amount float64 `json:"amount"`
}

type TransactionSplitRequest struct {
	UserID string  `json:"user_id"`
	Amount float64 `json:"amount"`
}

type CreateExpenseTransactionRequest struct {
	GroupID     string                    `json:"group_id"`
	Description string                    `json:"description"`
	Amount      float64                   `json:"amount"`
	Currency    string                    `json:"currency"`
	SplitType   string                    `json:"split_type"`
	Payers      []TransactionPayerRequest `json:"payers"` // Who paid money
	Splits      []TransactionSplitRequest `json:"splits"` // How it should be divided
	Category    string                    `json:"category"`
	Notes       string                    `json:"notes,omitempty"`
	IsCompleted bool                      `json:"is_completed,omitempty"`
}

func (r CreateExpenseTransactionRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.GroupID, validation.Required),
		validation.Field(&r.Description, validation.Required, validation.Length(1, 200)),
		validation.Field(&r.Amount, validation.Required, validation.Min(0.01)),
		validation.Field(&r.Currency, validation.Required, validation.Length(3, 3)),
		validation.Field(&r.SplitType, validation.Required, validation.In("equal", "exact", "percentage")),
		validation.Field(&r.Category, validation.Required),
		validation.Field(&r.Payers, validation.Required, validation.Length(1, 50)),
		validation.Field(&r.Splits, validation.Required, validation.Length(1, 50)),
	)
}

type UpdateTransactionRequest struct {
	Description string                    `json:"description,omitempty"`
	Amount      float64                   `json:"amount,omitempty"`
	SplitType   string                    `json:"split_type,omitempty"`
	Payers      []TransactionPayerRequest `json:"payers,omitempty"`
	Splits      []TransactionSplitRequest `json:"splits,omitempty"`
	Category    string                    `json:"category,omitempty"`
	Notes       string                    `json:"notes,omitempty"`
}

func (r UpdateTransactionRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Amount, validation.Min(0.0)),
		validation.Field(&r.SplitType, validation.In("", "equal", "exact", "percentage")),
	)
}

type CreateSettlementTransactionRequest struct {
	GroupID     string  `json:"group_id"`
	PayerID     string  `json:"payer_id"`
	PayeeID     string  `json:"payee_id"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Notes       string  `json:"notes,omitempty"`
	IsCompleted bool    `json:"is_completed,omitempty"`
}

func (r CreateSettlementTransactionRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.GroupID, validation.Required),
		validation.Field(&r.PayerID, validation.Required),
		validation.Field(&r.PayeeID, validation.Required),
		validation.Field(&r.Amount, validation.Required, validation.Min(0.01)),
		validation.Field(&r.Currency, validation.Required, validation.Length(3, 3)),
	)
}

type BulkSettlementsTransactionRequest struct {
	Settlements []CreateSettlementTransactionRequest `json:"settlements"`
}

func (r BulkSettlementsTransactionRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Settlements, validation.Required, validation.Length(1, 50)),
	)
}

type CompleteTransactionRequest struct {
	Notes            string `json:"notes,omitempty"`
	SettlementMethod string `json:"settlement_method,omitempty"`
	ProofOfPayment   string `json:"proof_of_payment,omitempty"`
}

func (r CompleteTransactionRequest) Validate() error {
	return validation.ValidateStruct(&r)
}


type PresignedURLRequest struct {
	FileName string `json:"file_name"`
}

func (r PresignedURLRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.FileName, validation.Required),
	)
}

type ConfirmUploadRequest struct {
	S3Key string `json:"s3_key"`
}

func (r ConfirmUploadRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.S3Key, validation.Required),
	)
}

type UpdateProfileRequest struct {
	Name string `json:"name"`
}

func (r UpdateProfileRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Name, validation.Required, validation.Length(3, 64)),
	)
}

type GoogleSignInRequest struct {
	IDToken string `json:"id_token"`
}

func (a GoogleSignInRequest) Validate() error {
	return validation.ValidateStruct(&a,
		validation.Field(&a.IDToken, validation.Required),
	)
}

type PushSubscriptionRequest struct {
	Endpoint string `json:"endpoint"`
	Keys     struct {
		P256dh string `json:"p256dh"`
		Auth   string `json:"auth"`
	} `json:"keys"`
}

func (r PushSubscriptionRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Endpoint, validation.Required, is.URL),
		validation.Field(&r.Keys.P256dh, validation.Required),
		validation.Field(&r.Keys.Auth, validation.Required),
	)
}
