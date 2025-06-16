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

type AddMemberToGroupRequest struct {
	UserID string `json:"user_id"`
}

func (r AddMemberToGroupRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.UserID, validation.Required),
	)
}

// Expense related requests
type ExpenseSplitRequest struct {
	UserID string  `json:"user_id"`
	Amount float64 `json:"amount"`
}

type CreateExpenseRequest struct {
	GroupID     string                `json:"group_id"`
	Description string                `json:"description"`
	Amount      float64               `json:"amount"`
	Currency    string                `json:"currency"`
	SplitType   string                `json:"split_type"`
	Splits      []ExpenseSplitRequest `json:"splits"`
	Category    string                `json:"category"`
	Notes       string                `json:"notes,omitempty"`
}

func (r CreateExpenseRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.GroupID, validation.Required),
		validation.Field(&r.Description, validation.Required, validation.Length(1, 200)),
		validation.Field(&r.Amount, validation.Required, validation.Min(0.01)),
		validation.Field(&r.Currency, validation.Required, validation.Length(3, 3)),
		validation.Field(&r.SplitType, validation.Required, validation.In("equal", "exact", "percentage")),
		validation.Field(&r.Category, validation.Required),
	)
}

type UpdateExpenseRequest struct {
	Description string                `json:"description,omitempty"`
	Amount      float64               `json:"amount,omitempty"`
	SplitType   string                `json:"split_type,omitempty"`
	Splits      []ExpenseSplitRequest `json:"splits,omitempty"`
	Category    string                `json:"category,omitempty"`
	Notes       string                `json:"notes,omitempty"`
}

func (r UpdateExpenseRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Amount, validation.Min(0.0)),                                  // Allow 0 for optional field
		validation.Field(&r.SplitType, validation.In("", "equal", "exact", "percentage")), // Allow empty
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

// Settlement related requests
type SettleDebtRequest struct {
	Notes string `json:"notes,omitempty"`
}

func (r SettleDebtRequest) Validate() error {
	return validation.ValidateStruct(&r)
}
