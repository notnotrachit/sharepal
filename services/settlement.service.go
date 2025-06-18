package services

import (
	"errors"
	"time"

	db "github.com/ebubekiryigit/golang-mongodb-rest-api-starter/models/db"
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// BalanceInfo represents the balance between two users
type BalanceInfo struct {
	UserID   primitive.ObjectID `json:"user_id"`
	UserName string             `json:"user_name"`
	Amount   float64            `json:"amount"`   // Positive means they owe money, negative means they are owed money
	Currency string             `json:"currency"` // Added currency field
}

// SettlementSuggestion represents a settlement suggestion with populated user data
type SettlementSuggestion struct {
	GroupID   primitive.ObjectID `json:"group_id"`
	PayerID   primitive.ObjectID `json:"payer_id"`
	PayerName string             `json:"payer_name"`
	PayeeID   primitive.ObjectID `json:"payee_id"`
	PayeeName string             `json:"payee_name"`
	Amount    float64            `json:"amount"`
	Currency  string             `json:"currency"`
	Status    string             `json:"status"`
}

// GroupBalance represents the overall balance summary for a group
type GroupBalance struct {
	GroupID   primitive.ObjectID `json:"group_id"`
	GroupName string             `json:"group_name"`
	Balances  []BalanceInfo      `json:"balances"`
}

func CalculateGroupBalances(groupID, userID primitive.ObjectID) (*GroupBalance, error) {
	// Check if user is group member
	group, err := GetGroupById(groupID, userID)
	if err != nil {
		return nil, err
	}

	// Get all expenses for the group
	var expenses []*db.Expense
	err = mgm.Coll(&db.Expense{}).SimpleFind(&expenses, bson.M{
		"group_id": groupID,
	})
	if err != nil {
		return nil, err
	}

	// Calculate balances
	balanceMap := make(map[primitive.ObjectID]float64)

	for _, expense := range expenses {
		// The payer is owed the full amount initially
		balanceMap[expense.PaidBy] += expense.Amount

		// Subtract each person's share
		for _, split := range expense.Splits {
			balanceMap[split.UserID] -= split.Amount
		}
	}

	// Get user details for the balances
	var users []*db.User
	err = mgm.Coll(&db.User{}).SimpleFind(&users, bson.M{
		"_id": bson.M{"$in": group.Members},
	})
	if err != nil {
		return nil, err
	}

	userMap := make(map[primitive.ObjectID]*db.User)
	for _, user := range users {
		userMap[user.ID] = user
	}

	// Convert to balance info for ALL group members
	var balances []BalanceInfo

	// Process ALL group members (including the requesting user)
	for _, memberID := range group.Members {
		user := userMap[memberID]
		if user == nil {
			continue
		}

		// Get balance for this member (0 if no expenses)
		balance := balanceMap[memberID]

		// Include ALL members, even with zero balance
		balances = append(balances, BalanceInfo{
			UserID:   memberID,
			UserName: user.Name,
			Amount:   balance,
			Currency: group.Currency, // Add currency field
		})
	}

	return &GroupBalance{
		GroupID:   groupID,
		GroupName: group.Name,
		Balances:  balances,
	}, nil
}

func CreateSettlement(groupID, payerID, payeeID primitive.ObjectID, amount float64, currency string, notes string) (*db.Settlement, error) {
	// Verify users are group members
	_, err := GetGroupById(groupID, payerID)
	if err != nil {
		return nil, err
	}

	_, err = GetGroupById(groupID, payeeID)
	if err != nil {
		return nil, err
	}

	settlement := db.NewSettlement(groupID, payerID, payeeID, amount, currency)
	settlement.Notes = notes

	err = mgm.Coll(settlement).Create(settlement)
	if err != nil {
		return nil, err
	}

	return settlement, nil
}

func MarkSettlementComplete(settlementID, userID primitive.ObjectID, notes string) error {
	settlement := &db.Settlement{}

	err := mgm.Coll(settlement).FindByID(settlementID, settlement)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("settlement not found")
		}
		return err
	}

	// Only the payer or payee can mark as complete
	if settlement.PayerID != userID && settlement.PayeeID != userID {
		return errors.New("only participants can mark settlement as complete")
	}

	if settlement.Status != db.SettlementPending {
		return errors.New("settlement is not pending")
	}

	now := time.Now()
	updateDoc := bson.M{
		"status":     db.SettlementCompleted,
		"settled_at": now,
	}

	if notes != "" {
		updateDoc["notes"] = notes
	}

	_, err = mgm.Coll(settlement).UpdateOne(mgm.Ctx(), bson.M{"_id": settlementID}, bson.M{
		"$set": updateDoc,
	})

	return err
}

func GetUserSettlements(userID primitive.ObjectID, status db.SettlementStatus) ([]*db.Settlement, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"payer_id": userID},
			{"payee_id": userID},
		},
	}

	if status != "" {
		filter["status"] = status
	}

	var settlements []*db.Settlement
	findOptions := options.Find().SetSort(bson.D{{"created_at", -1}})
	err := mgm.Coll(&db.Settlement{}).SimpleFind(&settlements, filter, findOptions)

	return settlements, err
}

func GetGroupSettlements(groupID, userID primitive.ObjectID) ([]*db.Settlement, error) {
	// Check if user is group member
	_, err := GetGroupById(groupID, userID)
	if err != nil {
		return nil, err
	}

	var settlements []*db.Settlement
	findOptions := options.Find().SetSort(bson.D{{"created_at", -1}})
	err = mgm.Coll(&db.Settlement{}).SimpleFind(&settlements, bson.M{
		"group_id": groupID,
	}, findOptions)

	return settlements, err
}

func SimplifyDebts(groupID, userID primitive.ObjectID) ([]SettlementSuggestion, error) {
	// Get group balances
	groupBalance, err := CalculateGroupBalances(groupID, userID)
	if err != nil {
		return nil, err
	}

	// Create a map of net balances
	netBalances := make(map[primitive.ObjectID]float64)
	for _, balance := range groupBalance.Balances {
		netBalances[balance.UserID] = balance.Amount
	}

	// Add the requesting user's balance (which is 0 relative to themselves)
	netBalances[userID] = 0

	// Calculate absolute balances
	for memberID := range netBalances {
		// Recalculate actual balance
		var expenses []*db.Expense
		err = mgm.Coll(&db.Expense{}).SimpleFind(&expenses, bson.M{
			"group_id": groupID,
		})
		if err != nil {
			return nil, err
		}

		balance := 0.0
		for _, expense := range expenses {
			if expense.PaidBy == memberID {
				balance += expense.Amount
			}
			for _, split := range expense.Splits {
				if split.UserID == memberID {
					balance -= split.Amount
				}
			}
		}
		netBalances[memberID] = balance
	}

	// Simplify debts using a greedy algorithm
	// Sign convention: Positive balance = owes money, Negative balance = is owed money
	var settlements []SettlementSuggestion

	// Get all users for name lookup
	var allUsers []*db.User
	userIDs := make([]primitive.ObjectID, 0, len(netBalances))
	for userID := range netBalances {
		userIDs = append(userIDs, userID)
	}

	err = mgm.Coll(&db.User{}).SimpleFind(&allUsers, bson.M{
		"_id": bson.M{"$in": userIDs},
	})
	if err != nil {
		return nil, err
	}

	// Create user lookup map
	userLookup := make(map[primitive.ObjectID]*db.User)
	for _, user := range allUsers {
		userLookup[user.ID] = user
	}

	for {
		// Find the person who owes the most (most positive balance)
		var maxDebtorID primitive.ObjectID
		maxDebt := 0.0

		// Find the person who is owed the most (most negative balance)
		var maxCreditorID primitive.ObjectID
		maxCredit := 0.0

		for userID, balance := range netBalances {
			if balance > maxDebt {
				maxDebt = balance
				maxDebtorID = userID // Most positive = owes most money
			}
			if balance < maxCredit {
				maxCredit = balance
				maxCreditorID = userID // Most negative = is owed most money
			}
		}

		// If no significant debt remains, break
		if maxDebt <= 0.01 && maxCredit >= -0.01 {
			break
		}

		// Calculate settlement amount
		settlementAmount := maxDebt
		if -maxCredit < maxDebt {
			settlementAmount = -maxCredit
		}

		if settlementAmount > 0.01 {
			// Create settlement suggestion with populated user data
			group, _ := GetGroupById(groupID, userID)

			payerName := "Unknown User"
			payeeName := "Unknown User"

			if payer, exists := userLookup[maxDebtorID]; exists {
				payerName = payer.Name
			}
			if payee, exists := userLookup[maxCreditorID]; exists {
				payeeName = payee.Name
			}

			settlement := SettlementSuggestion{
				GroupID:   groupID,
				PayerID:   maxDebtorID,
				PayerName: payerName,
				PayeeID:   maxCreditorID,
				PayeeName: payeeName,
				Amount:    settlementAmount,
				Currency:  group.Currency,
				Status:    string(db.SettlementPending),
			}
			settlements = append(settlements, settlement)

			// Update balances
			netBalances[maxDebtorID] -= settlementAmount
			netBalances[maxCreditorID] += settlementAmount
		} else {
			break
		}
	}

	return settlements, nil
}
