package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/ebubekiryigit/golang-mongodb-rest-api-starter/models"
	db "github.com/ebubekiryigit/golang-mongodb-rest-api-starter/models/db"
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TransactionService handles all transaction operations with atomic balance updates
type TransactionService struct{}

// CreateExpenseTransaction creates a new expense and updates all related balances atomically
func (ts *TransactionService) CreateExpenseTransaction(userID primitive.ObjectID, req models.CreateExpenseTransactionRequest) (*db.Transaction, error) {
	groupID, err := primitive.ObjectIDFromHex(req.GroupID)
	if err != nil {
		return nil, errors.New("invalid group ID")
	}

	// Check if user is group member
	group, err := GetGroupById(groupID, userID)
	if err != nil {
		return nil, err
	}

	// Create the transaction
	transaction := db.NewExpenseTransaction(groupID, req.Description, req.Amount, req.Currency, userID, db.SplitType(req.SplitType), req.Category)
	transaction.Notes = req.Notes
	transaction.IsCompleted = req.IsCompleted

	// Process payers and splits
	if len(req.Payers) == 0 {
		return nil, errors.New("at least one payer is required")
	}
	if len(req.Splits) == 0 {
		return nil, errors.New("at least one split is required")
	}

	// Validate total amounts
	var totalPaid, totalSplit float64
	for _, payer := range req.Payers {
		totalPaid += payer.Amount
	}
	for _, split := range req.Splits {
		totalSplit += split.Amount
	}

	if math.Abs(totalPaid-req.Amount) > 0.01 {
		return nil, errors.New("total paid amount must equal transaction amount")
	}
	if math.Abs(totalSplit-req.Amount) > 0.01 {
		return nil, errors.New("total split amount must equal transaction amount")
	}

	// Process payers
	for _, payer := range req.Payers {
		payerUserID, err := primitive.ObjectIDFromHex(payer.UserID)
		if err != nil {
			return nil, errors.New("invalid payer user ID")
		}

		// Get user name
		user, err := FindUserById(payerUserID)
		if err != nil {
			return nil, errors.New("payer user not found")
		}

		transaction.Payers = append(transaction.Payers, db.TransactionPayer{
			UserID:   payerUserID,
			UserName: user.Name,
			Amount:   payer.Amount,
		})
	}

	// Process splits
	for _, split := range req.Splits {
		splitUserID, err := primitive.ObjectIDFromHex(split.UserID)
		if err != nil {
			return nil, errors.New("invalid split user ID")
		}

		// Get user name
		user, err := FindUserById(splitUserID)
		if err != nil {
			return nil, errors.New("split user not found")
		}

		transaction.Splits = append(transaction.Splits, db.TransactionSplit{
			UserID:   splitUserID,
			UserName: user.Name,
			Amount:   split.Amount,
		})
	}

	// Calculate net participants (for balance updates)
	participantMap := make(map[primitive.ObjectID]*db.TransactionParticipant)

	// Add payers (positive amounts - they paid money)
	for _, payer := range transaction.Payers {
		if participant, exists := participantMap[payer.UserID]; exists {
			participant.Amount += payer.Amount
		} else {
			participantMap[payer.UserID] = &db.TransactionParticipant{
				UserID:    payer.UserID,
				UserName:  payer.UserName,
				Amount:    payer.Amount,
				ShareType: "payer",
			}
		}
	}

	// Subtract splits (negative amounts - they owe money)
	for _, split := range transaction.Splits {
		if participant, exists := participantMap[split.UserID]; exists {
			participant.Amount -= split.Amount
			if participant.ShareType == "payer" {
				participant.ShareType = "both"
			} else {
				participant.ShareType = "split"
			}
		} else {
			participantMap[split.UserID] = &db.TransactionParticipant{
				UserID:    split.UserID,
				UserName:  split.UserName,
				Amount:    -split.Amount,
				ShareType: "split",
			}
		}
	}

	// Convert map to slice
	for _, participant := range participantMap {
		transaction.Participants = append(transaction.Participants, *participant)
	}

	// Execute transaction with balance updates atomically
	return ts.executeTransactionWithBalanceUpdate(transaction, group)
}

// CreateSettlementTransaction creates a settlement between two users
func (ts *TransactionService) CreateSettlementTransaction(groupID, payerID, payeeID primitive.ObjectID, amount float64, currency string, notes string, isCompleted bool, createdBy primitive.ObjectID) (*db.Transaction, error) {
	// Verify users are group members
	group, err := GetGroupById(groupID, createdBy)
	if err != nil {
		return nil, err
	}

	// Get user names for denormalization
	payer, err := FindUserById(payerID)
	if err != nil {
		return nil, errors.New("payer not found")
	}

	payee, err := FindUserById(payeeID)
	if err != nil {
		return nil, errors.New("payee not found")
	}

	transaction := db.NewSettlementTransaction(groupID, payerID, payeeID, amount, currency)
	transaction.Notes = notes
	transaction.IsCompleted = isCompleted
	transaction.CreatedBy = createdBy

	// Update participant names
	for i := range transaction.Participants {
		if transaction.Participants[i].UserID == payerID {
			transaction.Participants[i].UserName = payer.Name
		} else if transaction.Participants[i].UserID == payeeID {
			transaction.Participants[i].UserName = payee.Name
		}
	}

	return ts.executeTransactionWithBalanceUpdate(transaction, group)
}

// executeTransactionWithBalanceUpdate performs atomic transaction creation and balance updates
func (ts *TransactionService) executeTransactionWithBalanceUpdate(transaction *db.Transaction, group *db.Group) (*db.Transaction, error) {
	_, client, _, err := mgm.DefaultConfigs()
	if err != nil {
		return nil, err
	}

	session, err := client.StartSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(context.Background())

	var result *db.Transaction

	err = mongo.WithSession(context.Background(), session, func(sc mongo.SessionContext) error {
		// 1. Create the transaction
		if err := mgm.Coll(transaction).Create(transaction); err != nil {
			return err
		}

		// 2. Update balances for expenses using payers and splits directly
		if transaction.Type == db.TransactionTypeExpense {
			// Update balances for payers (they paid money)
			for _, payer := range transaction.Payers {
				if err := ts.updateUserBalance(sc, transaction.GroupID, payer.UserID, payer.UserName, payer.Amount, 0, transaction.Type, transaction.ID, group.Currency); err != nil {
					return err
				}
			}

			// Update balances for splits (they owe money)
			for _, split := range transaction.Splits {
				if err := ts.updateUserBalance(sc, transaction.GroupID, split.UserID, split.UserName, 0, split.Amount, transaction.Type, transaction.ID, group.Currency); err != nil {
					return err
				}
			}
		} else {
			// For settlements, handle payer and payee correctly
			for _, participant := range transaction.Participants {
				if participant.Amount > 0 {
					// They paid money (payer) - increase their TotalPaid
					if err := ts.updateUserBalance(sc, transaction.GroupID, participant.UserID, participant.UserName, participant.Amount, 0, transaction.Type, transaction.ID, group.Currency); err != nil {
						return err
					}
				} else {
					// They received money (payee) - increase their TotalOwed to effectively reduce their debt
					// Convert negative amount to positive for received amount
					receivedAmount := -participant.Amount
					if err := ts.updateUserBalance(sc, transaction.GroupID, participant.UserID, participant.UserName, 0, receivedAmount, transaction.Type, transaction.ID, group.Currency); err != nil {
						return err
					}
				}
			}
		}

		result = transaction
		return nil
	})

	// Send notifications in background after successful transaction
	if err == nil {
		go ts.sendTransactionNotifications(transaction, group)
	}

	return result, err
}

// updateUserBalance updates a user's balance in a group atomically
func (ts *TransactionService) updateUserBalance(sc mongo.SessionContext, groupID, userID primitive.ObjectID, userName string, amountPaid, amountOwed float64, transactionType db.TransactionType, transactionID primitive.ObjectID, currency string) error {
	balance := &db.GroupBalance{}

	// Try to find existing balance
	filter := bson.M{
		"group_id": groupID,
		"user_id":  userID,
	}

	err := mgm.Coll(balance).FindOne(mgm.Ctx(), filter).Decode(balance)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// Create new balance record
			balance = db.NewGroupBalance(groupID, userID, userName, currency)
		} else {
			return err
		}
	}
	// Update balance with the provided amounts
	balance.UpdateBalance(amountPaid, amountOwed, transactionID)

	// Upsert the balance record
	opts := options.Replace().SetUpsert(true)
	_, err = mgm.Coll(balance).ReplaceOne(sc, filter, balance, opts)
	return err
}

// GetGroupBalances returns current balances for all group members
func (ts *TransactionService) GetGroupBalances(groupID, userID primitive.ObjectID) ([]*db.GroupBalance, error) {
	// Check if user is group member
	_, err := GetGroupById(groupID, userID)
	if err != nil {
		return nil, err
	}

	var balances []*db.GroupBalance
	err = mgm.Coll(&db.GroupBalance{}).SimpleFind(&balances, bson.M{
		"group_id": groupID,
	})

	return balances, err
}

// logGroupTransactions logs all transactions for debugging purposes
func (ts *TransactionService) logGroupTransactions(groupID primitive.ObjectID) {
	var transactions []*db.Transaction
	err := mgm.Coll(&db.Transaction{}).SimpleFind(&transactions, bson.M{
		"group_id": groupID,
	})

	if err != nil {
		return
	}
}

// SimplifyDebtsFromBalances calculates optimal settlements from current balances
func (ts *TransactionService) SimplifyDebtsFromBalances(groupID, userID primitive.ObjectID) ([]models.SettlementSuggestion, error) {
	balances, err := ts.GetGroupBalances(groupID, userID)
	if err != nil {
		// Log the error for debugging
		return nil, errors.New("failed to get group balances: " + err.Error())
	}

	if len(balances) == 0 {
		return []models.SettlementSuggestion{}, nil
	}

	// Debug: Log recent transactions to understand how we got these balances
	ts.logGroupTransactions(groupID)

	// Convert to map for simplification algorithm
	netBalances := make(map[primitive.ObjectID]float64)
	userLookup := make(map[primitive.ObjectID]string)
	currency := balances[0].Currency

	for _, balance := range balances {
		netBalances[balance.UserID] = balance.Balance
		userLookup[balance.UserID] = balance.UserName
	}

	// Use the same simplification algorithm but with pre-calculated balances
	var settlements []models.SettlementSuggestion


	for {
		var maxCreditorID, maxDebtorID primitive.ObjectID
		maxCredit, maxDebt := 0.0, 0.0

		for userID, balance := range netBalances {
			if balance > maxCredit {
				maxCredit = balance
				maxCreditorID = userID
			}
			if balance < maxDebt {
				maxDebt = balance
				maxDebtorID = userID
			}
		}



		// maxCredit > 0 means user is owed money
		// maxDebt < 0 means user owes money
		if maxCredit <= 0.01 && maxDebt >= -0.01 {
			break
		}

		settlementAmount := maxCredit
		if -maxDebt < maxCredit {
			settlementAmount = -maxDebt
		}

		if settlementAmount > 0.01 {
			settlement := models.SettlementSuggestion{
				GroupID:   groupID,
				PayerID:   maxDebtorID, // Person who owes money (negative balance)
				PayerName: userLookup[maxDebtorID],
				PayeeID:   maxCreditorID, // Person who is owed money (positive balance)
				PayeeName: userLookup[maxCreditorID],
				Amount:    settlementAmount,
				Currency:  currency,
				Status:    "pending",
			}
			settlements = append(settlements, settlement)

			netBalances[maxDebtorID] += settlementAmount   // Debtor's balance increases (less negative)
			netBalances[maxCreditorID] -= settlementAmount // Creditor's balance decreases (less positive)
		} else {
			break
		}
	}

	return settlements, nil
}

// MarkTransactionComplete marks a settlement transaction as completed
func (ts *TransactionService) MarkTransactionComplete(transactionID, userID primitive.ObjectID, notes string, settlementMethod string, proofOfPayment string) error {
	transaction := &db.Transaction{}

	err := mgm.Coll(transaction).FindByID(transactionID, transaction)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("transaction not found")
		}
		return err
	}

	if transaction.Type != db.TransactionTypeSettlement {
		return errors.New("only settlement transactions can be marked as complete")
	}

	// Only participants can mark as complete
	isParticipant := false
	for _, participant := range transaction.Participants {
		if participant.UserID == userID {
			isParticipant = true
			break
		}
	}

	if !isParticipant {
		return errors.New("only participants can mark settlement as complete")
	}

	if transaction.IsCompleted {
		return errors.New("transaction is already completed")
	}

	now := time.Now()
	updateDoc := bson.M{
		"is_completed": true,
		"settled_at":   now,
		"updated_at":   now,
		"updated_by":   userID,
	}

	if notes != "" {
		updateDoc["notes"] = notes
	}
	if settlementMethod != "" {
		updateDoc["settlement_method"] = settlementMethod
	}
	if proofOfPayment != "" {
		updateDoc["proof_of_payment"] = proofOfPayment
	}

	_, err = mgm.Coll(transaction).UpdateOne(mgm.Ctx(), bson.M{"_id": transactionID}, bson.M{
		"$set": updateDoc,
	})

	return err
}

// GetGroupTransactions returns all transactions for a group
func (ts *TransactionService) GetGroupTransactions(groupID, userID primitive.ObjectID, transactionType string, page, limit int) ([]*db.Transaction, error) {
	// Check if user is group member
	_, err := GetGroupById(groupID, userID)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"group_id": groupID}
	if transactionType != "" {
		filter["type"] = transactionType
	}

	var transactions []*db.Transaction
	findOptions := options.Find().
		SetSkip(int64(page * limit)).
		SetLimit(int64(limit + 1)).
		SetSort(bson.D{{Key: "date", Value: -1}})

	err = mgm.Coll(&db.Transaction{}).SimpleFind(&transactions, filter, findOptions)
	if err != nil {
		return nil, err
	}

	// Collect all unique user IDs first
	userIDs := make(map[primitive.ObjectID]bool)
	for _, transaction := range transactions {
		userIDs[transaction.CreatedBy] = true
		for _, participant := range transaction.Participants {
			userIDs[participant.UserID] = true
		}
		for _, payer := range transaction.Payers {
			userIDs[payer.UserID] = true
		}
		for _, split := range transaction.Splits {
			userIDs[split.UserID] = true
		}
	}

	userProfilePics := make(map[primitive.ObjectID]string)
	for userID := range userIDs {
		user, err := GetUserWithProfilePictureURL(userID, 60)
		if err == nil {
			userProfilePics[userID] = user.ProfilePicUrl
		}
	}

	for _, transaction := range transactions {
		if profilePicUrl, exists := userProfilePics[transaction.CreatedBy]; exists {
			transaction.CreatorProfilePicUrl = profilePicUrl
		}

		for i, participant := range transaction.Participants {
			if profilePicUrl, exists := userProfilePics[participant.UserID]; exists {
				transaction.Participants[i].ProfilePicUrl = profilePicUrl
			}
		}

		for i, payer := range transaction.Payers {
			if profilePicUrl, exists := userProfilePics[payer.UserID]; exists {
				transaction.Payers[i].ProfilePicUrl = profilePicUrl
			}
		}

		for i, split := range transaction.Splits {
			if profilePicUrl, exists := userProfilePics[split.UserID]; exists {
				transaction.Splits[i].ProfilePicUrl = profilePicUrl
			}
		}
	}

	return transactions, err
}

// GetTransactionById returns a single transaction by ID
func (ts *TransactionService) GetTransactionById(transactionID, userID primitive.ObjectID) (*db.Transaction, error) {
	transaction := &db.Transaction{}

	err := mgm.Coll(transaction).FindByID(transactionID, transaction)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("transaction not found")
		}
		return nil, err
	}

	// Check if user is group member
	_, err = GetGroupById(transaction.GroupID, userID)
	if err != nil {
		return nil, err
	}

	userIDs := make(map[primitive.ObjectID]bool)
	userIDs[transaction.CreatedBy] = true
	for _, participant := range transaction.Participants {
		userIDs[participant.UserID] = true
	}
	for _, payer := range transaction.Payers {
		userIDs[payer.UserID] = true
	}
	for _, split := range transaction.Splits {
		userIDs[split.UserID] = true
	}

	userProfilePics := make(map[primitive.ObjectID]string)
	for userID := range userIDs {
		user, err := GetUserWithProfilePictureURL(userID, 60)
		if err == nil {
			userProfilePics[userID] = user.ProfilePicUrl
		}
	}

	if profilePicUrl, exists := userProfilePics[transaction.CreatedBy]; exists {
		transaction.CreatorProfilePicUrl = profilePicUrl
	}

	for i, participant := range transaction.Participants {
		if profilePicUrl, exists := userProfilePics[participant.UserID]; exists {
			transaction.Participants[i].ProfilePicUrl = profilePicUrl
		}
	}

	for i, payer := range transaction.Payers {
		if profilePicUrl, exists := userProfilePics[payer.UserID]; exists {
			transaction.Payers[i].ProfilePicUrl = profilePicUrl
		}
	}

	for i, split := range transaction.Splits {
		if profilePicUrl, exists := userProfilePics[split.UserID]; exists {
			transaction.Splits[i].ProfilePicUrl = profilePicUrl
		}
	}

	return transaction, nil
}

// UpdateTransaction updates an expense transaction
func (ts *TransactionService) UpdateTransaction(transactionID, userID primitive.ObjectID, req models.UpdateTransactionRequest) error {
	transaction, err := ts.GetTransactionById(transactionID, userID)
	if err != nil {
		return err
	}

	if transaction.Type != db.TransactionTypeExpense {
		return errors.New("only expense transactions can be updated")
	}

	// Only the creator can update the transaction
	if transaction.CreatedBy != userID {
		return errors.New("only the creator can update this transaction")
	}

	if transaction.IsCompleted {
		return errors.New("completed transactions cannot be updated")
	}

	// Create update document
	updateDoc := bson.M{
		"updated_at": time.Now(),
		"updated_by": userID,
	}

	if req.Description != "" {
		updateDoc["description"] = req.Description
	}
	if req.Amount > 0 {
		updateDoc["amount"] = req.Amount
	}
	if req.Category != "" {
		updateDoc["category"] = req.Category
	}
	if req.Notes != "" {
		updateDoc["notes"] = req.Notes
	}

	// Handle split updates with balance recalculation
	if req.SplitType != "" && (len(req.Payers) > 0 || len(req.Splits) > 0) {
		// This would require complex balance recalculation
		// For now, prevent updates that change payers/splits
		return errors.New("payer/split updates not yet supported - please delete and recreate the transaction")
	}

	if len(updateDoc) <= 2 { // Only updated_at and updated_by
		return errors.New("no fields to update")
	}

	_, err = mgm.Coll(transaction).UpdateOne(mgm.Ctx(), bson.M{"_id": transactionID}, bson.M{
		"$set": updateDoc,
	})

	return err
}

// DeleteTransaction deletes a transaction and recalculates balances
func (ts *TransactionService) DeleteTransaction(transactionID, userID primitive.ObjectID) error {
	transaction, err := ts.GetTransactionById(transactionID, userID)
	if err != nil {
		return err
	}

	// Only the creator can delete the transaction
	if transaction.CreatedBy != userID {
		return errors.New("only the creator can delete this transaction")
	}

	if transaction.IsCompleted && transaction.Type == db.TransactionTypeSettlement {
		return errors.New("completed settlements cannot be deleted")
	}

	// Get group for currency info
	group, err := GetGroupById(transaction.GroupID, userID)
	if err != nil {
		return err
	}

	// Start transaction to delete and update balances atomically
	_, client, _, err := mgm.DefaultConfigs()
	if err != nil {
		return err
	}

	session, err := client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(context.Background())

	err = mongo.WithSession(context.Background(), session, func(sc mongo.SessionContext) error {
		// Delete the transaction
		if err := mgm.Coll(transaction).Delete(transaction); err != nil {
			return err
		}

		// Reverse balance updates for all participants
		for _, participant := range transaction.Participants {
			if err := ts.reverseUserBalance(sc, transaction.GroupID, participant.UserID, participant.Amount, transaction.Type, group.Currency); err != nil {
				return err
			}
		}

		return nil
	})

	return err
}

// reverseUserBalance reverses a balance update when a transaction is deleted
func (ts *TransactionService) reverseUserBalance(sc mongo.SessionContext, groupID, userID primitive.ObjectID, amount float64, transactionType db.TransactionType, currency string) error {
	balance := &db.GroupBalance{}

	filter := bson.M{
		"group_id": groupID,
		"user_id":  userID,
	}

	err := mgm.Coll(balance).FindOne(mgm.Ctx(), filter).Decode(balance)
	if err != nil {
		return err // Balance should exist if transaction existed
	}

	// Reverse the balance changes
	var amountPaid, amountOwed float64

	switch transactionType {
	case db.TransactionTypeExpense:
		if amount > 0 {
			amountPaid = -amount // Reverse payment
		} else {
			amountOwed = amount // Reverse debt (negative amount becomes positive)
		}
	case db.TransactionTypeSettlement:
		if amount > 0 {
			amountPaid = -amount // Reverse payment
		} else {
			amountOwed = -amount // Reverse receipt
		}
	}

	// Update balance
	balance.TotalPaid += amountPaid
	balance.TotalOwed += amountOwed
	balance.Balance = balance.TotalPaid - balance.TotalOwed
	balance.LastUpdated = time.Now()
	balance.Version++

	// Update the balance record
	_, err = mgm.Coll(balance).ReplaceOne(sc, filter, balance)
	return err
}

// GetGroupBalanceHistory returns balance change history for a group
func (ts *TransactionService) GetGroupBalanceHistory(groupID, userID primitive.ObjectID, days int) (interface{}, error) {
	// Check if user is group member
	_, err := GetGroupById(groupID, userID)
	if err != nil {
		return nil, err
	}

	// Get transactions from the last N days
	startDate := time.Now().AddDate(0, 0, -days)

	var transactions []*db.Transaction
	filter := bson.M{
		"group_id": groupID,
		"date":     bson.M{"$gte": startDate},
	}

	findOptions := options.Find().SetSort(bson.D{{Key: "date", Value: 1}})
	err = mgm.Coll(&db.Transaction{}).SimpleFind(&transactions, filter, findOptions)
	if err != nil {
		return nil, err
	}

	// Build balance history by replaying transactions
	history := make(map[string][]map[string]interface{})

	for _, transaction := range transactions {
		dateStr := transaction.Date.Format("2006-01-02")
		if history[dateStr] == nil {
			history[dateStr] = []map[string]interface{}{}
		}

		history[dateStr] = append(history[dateStr], map[string]interface{}{
			"transaction_id": transaction.ID,
			"type":           transaction.Type,
			"description":    transaction.Description,
			"amount":         transaction.Amount,
			"participants":   transaction.Participants,
		})
	}

	return history, nil
}

// GetGroupAnalytics returns analytics data for a group
func (ts *TransactionService) GetGroupAnalytics(groupID, userID primitive.ObjectID) (interface{}, error) {
	// Check if user is group member
	group, err := GetGroupById(groupID, userID)
	if err != nil {
		return nil, err
	}

	// Get all transactions for the group
	var transactions []*db.Transaction
	err = mgm.Coll(&db.Transaction{}).SimpleFind(&transactions, bson.M{
		"group_id": groupID,
	})
	if err != nil {
		return nil, err
	}

	// Get all balances for the group
	balances, err := ts.GetGroupBalances(groupID, userID)
	if err != nil {
		return nil, err
	}

	// Calculate analytics
	analytics := map[string]interface{}{
		"group_id":           groupID,
		"group_name":         group.Name,
		"total_transactions": len(transactions),
		"total_expenses":     0,
		"total_settlements":  0,
		"total_amount":       0.0,
		"currency":           group.Currency,
		"member_count":       len(group.Members),
		"balances_summary": map[string]int{
			"positive": 0, // Members who are owed money
			"negative": 0, // Members who owe money
			"zero":     0, // Members with zero balance
		},
	}

	// Process transactions
	var totalExpenseAmount, totalSettlementAmount float64
	expenseCount, settlementCount := 0, 0

	for _, transaction := range transactions {
		switch transaction.Type {
		case db.TransactionTypeExpense:
			expenseCount++
			totalExpenseAmount += transaction.Amount
		case db.TransactionTypeSettlement:
			settlementCount++
			totalSettlementAmount += transaction.Amount
		}
	}

	analytics["total_expenses"] = expenseCount
	analytics["total_settlements"] = settlementCount
	analytics["total_expense_amount"] = totalExpenseAmount
	analytics["total_settlement_amount"] = totalSettlementAmount
	analytics["total_amount"] = totalExpenseAmount

	// Process balances
	balanceSummary := analytics["balances_summary"].(map[string]int)
	for _, balance := range balances {
		if balance.Balance > 0.01 {
			balanceSummary["positive"]++
		} else if balance.Balance < -0.01 {
			balanceSummary["negative"]++
		} else {
			balanceSummary["zero"]++
		}
	}

	return analytics, nil
}

// CreateBulkSettlements creates multiple settlements from suggested settlements
func (ts *TransactionService) CreateBulkSettlements(groupID, userID primitive.ObjectID, settlementRequests []models.CreateSettlementTransactionRequest) ([]*db.Transaction, error) {
	// Verify user is group member
	_, err := GetGroupById(groupID, userID)
	if err != nil {
		return nil, err
	}

	if len(settlementRequests) == 0 {
		return nil, errors.New("no settlements provided")
	}

	if len(settlementRequests) > 50 {
		return nil, errors.New("too many settlements - maximum 50 per request")
	}

	var settlements []*db.Transaction

	// Create settlements one by one (could be optimized with bulk operations)
	for _, req := range settlementRequests {
		if req.GroupID != groupID.Hex() {
			return nil, errors.New("all settlements must be for the same group")
		}

		payerID, err := primitive.ObjectIDFromHex(req.PayerID)
		if err != nil {
			return nil, errors.New("invalid payer ID in settlement request")
		}

		payeeID, err := primitive.ObjectIDFromHex(req.PayeeID)
		if err != nil {
			return nil, errors.New("invalid payee ID in settlement request")
		}

		settlement, err := ts.CreateSettlementTransaction(groupID, payerID, payeeID, req.Amount, req.Currency, req.Notes, req.IsCompleted, userID)
		if err != nil {
			return nil, err
		}

		settlements = append(settlements, settlement)
	}

	return settlements, nil
}

// RecalculateGroupBalances recalculates all balances for a group (admin operation)
func (ts *TransactionService) RecalculateGroupBalances(groupID, userID primitive.ObjectID) error {
	// Check if user is group member
	group, err := GetGroupById(groupID, userID)
	if err != nil {
		return err
	}

	// Get all transactions for the group
	var transactions []*db.Transaction
	findOptions := options.Find().SetSort(bson.D{{Key: "date", Value: 1}})
	err = mgm.Coll(&db.Transaction{}).SimpleFind(&transactions, bson.M{
		"group_id": groupID,
	}, findOptions)
	if err != nil {
		return err
	}

	// Reset all balances for the group
	_, err = mgm.Coll(&db.GroupBalance{}).DeleteMany(mgm.Ctx(), bson.M{
		"group_id": groupID,
	})
	if err != nil {
		return err
	}

	// Replay all transactions to rebuild balances
	for _, transaction := range transactions {
		for _, participant := range transaction.Participants {
			var amountPaid, amountOwed float64

			switch transaction.Type {
			case db.TransactionTypeExpense:
				if participant.Amount > 0 {
					amountPaid = participant.Amount
				} else {
					amountOwed = -participant.Amount
				}
			case db.TransactionTypeSettlement:
				if participant.Amount > 0 {
					amountPaid = participant.Amount
				} else {
					amountOwed = participant.Amount // This will be negative
				}
			}

			// Update or create balance
			filter := bson.M{
				"group_id": groupID,
				"user_id":  participant.UserID,
			}

			balance := &db.GroupBalance{}
			err := mgm.Coll(balance).FindOne(mgm.Ctx(), filter).Decode(balance)
			if err != nil {
				if err == mongo.ErrNoDocuments {
					// Create new balance
					balance = db.NewGroupBalance(groupID, participant.UserID, participant.UserName, group.Currency)
				} else {
					return err
				}
			}

			balance.UpdateBalance(amountPaid, amountOwed, transaction.ID)

			// Upsert the balance
			opts := options.Replace().SetUpsert(true)
			_, err = mgm.Coll(balance).ReplaceOne(mgm.Ctx(), filter, balance, opts)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// GetUserTransactions returns all transactions for a user across all groups
func (ts *TransactionService) GetUserTransactions(userID primitive.ObjectID, transactionType string, page, limit int) ([]*db.Transaction, error) {
	filter := bson.M{
		"participants.user_id": userID,
	}

	if transactionType != "" {
		filter["type"] = transactionType
	}

	var transactions []*db.Transaction
	findOptions := options.Find().
		SetSkip(int64(page * limit)).
		SetLimit(int64(limit + 1)).
		SetSort(bson.D{{Key: "date", Value: -1}})

	err := mgm.Coll(&db.Transaction{}).SimpleFind(&transactions, filter, findOptions)
	return transactions, err
}

// GetUserBalances returns all group balances for a user
func (ts *TransactionService) GetUserBalances(userID primitive.ObjectID) ([]*db.GroupBalance, error) {
	var balances []*db.GroupBalance
	err := mgm.Coll(&db.GroupBalance{}).SimpleFind(&balances, bson.M{
		"user_id": userID,
	})
	return balances, err
}

// GetUserAnalytics returns analytics data for a user
func (ts *TransactionService) GetUserAnalytics(userID primitive.ObjectID) (interface{}, error) {
	// Get user's balances across all groups
	balances, err := ts.GetUserBalances(userID)
	if err != nil {
		return nil, err
	}

	// Get user's transactions across all groups
	transactions, err := ts.GetUserTransactions(userID, "", 0, 1000) // Get up to 1000 recent transactions
	if err != nil {
		return nil, err
	}

	// Calculate analytics
	analytics := map[string]interface{}{
		"user_id":            userID,
		"total_groups":       len(balances),
		"total_transactions": len(transactions),
		"total_expenses":     0,
		"total_settlements":  0,
		"net_balance":        0.0,
		"groups_summary": map[string]int{
			"owe_money":  0, // Groups where user owes money
			"owed_money": 0, // Groups where user is owed money
			"balanced":   0, // Groups where user has zero balance
		},
	}

	// Process balances
	var netBalance float64
	groupsSummary := analytics["groups_summary"].(map[string]int)

	for _, balance := range balances {
		netBalance += balance.Balance

		if balance.Balance > 0.01 {
			groupsSummary["owed_money"]++
		} else if balance.Balance < -0.01 {
			groupsSummary["owe_money"]++
		} else {
			groupsSummary["balanced"]++
		}
	}

	analytics["net_balance"] = netBalance

	// Process transactions
	expenseCount, settlementCount := 0, 0
	for _, transaction := range transactions {
		switch transaction.Type {
		case db.TransactionTypeExpense:
			expenseCount++
		case db.TransactionTypeSettlement:
			settlementCount++
		}
	}

	analytics["total_expenses"] = expenseCount
	analytics["total_settlements"] = settlementCount

	return analytics, nil
}

// sendTransactionNotifications sends notifications to all transaction participants in background
func (ts *TransactionService) sendTransactionNotifications(transaction *db.Transaction, group *db.Group) {
	for _, participant := range transaction.Participants {
		// Skip sending notification to the transaction creator
		// if participant.UserID == transaction.CreatedBy {
		// 	continue
		// }

		user, err := FindUserById(participant.UserID)
		if err != nil {
			log.Printf("Error finding user %s for notification: %v\n", participant.UserID.Hex(), err)
			continue
		}

		// // Send FCM notification if user has FCM token
		// if user.FCMToken != "" {
		// 	var title, body string
		// 	notificationData := map[string]string{
		// 		"type":           "transaction",
		// 		"transaction_id": transaction.ID.Hex(),
		// 		"group_id":       group.ID.Hex(),
		// 	}

		// 	switch transaction.Type {
		// 	case db.TransactionTypeExpense:
		// 		title = "New Expense Added"
		// 		body = fmt.Sprintf("A new expense '%s' was added to %s", transaction.Description, group.Name)
		// 		notificationData["expense_amount"] = fmt.Sprintf("%.2f", transaction.Amount)
		// 	case db.TransactionTypeSettlement:
		// 		title = "New Settlement"
		// 		body = fmt.Sprintf("A settlement was recorded in %s", group.Name)
		// 		notificationData["settlement_amount"] = fmt.Sprintf("%.2f", transaction.Amount)
		// 	}

		// 	err := SendFCMNotification(user.FCMToken, title, body, notificationData)
		// 	if err != nil {
		// 		log.Printf("Error sending FCM notification to %s: %v\n", user.Email, err)
		// 	}
		// }

		// Send web push notification if user has subscriptions
		subs, err := GetPushSubscriptionsByUserID(user.ID)
		if err != nil {
			log.Printf("Error getting push subscriptions for user %s: %v\n", user.Email, err)
			continue
		}

		if len(subs) > 0 {
			var title, body string
			switch transaction.Type {
			case db.TransactionTypeExpense:
				title = "New Expense Added"
				body = fmt.Sprintf("A new expense '%s' was added to %s", transaction.Description, group.Name)
			case db.TransactionTypeSettlement:
				title = "New Settlement"
				body = fmt.Sprintf("A settlement was recorded in %s", group.Name)
			}

			notificationData := map[string]interface{}{
				"title": title,
				"body":  body,
				"data": map[string]interface{}{
					"type":           "transaction",
					"transaction_id": transaction.ID.Hex(),
					"group_id":       group.ID.Hex(),
					"amount":         transaction.Amount,
					"currency":       transaction.Currency,
				},
			}

			for _, sub := range subs {
				err := Notification.SendJSONNotification(sub, notificationData)
				if err != nil {
					log.Printf("Error sending web push notification to %s: %v\n", user.Email, err)
				}
				log.Default().Printf("Sent web push notification to %s for transaction %s\n", user.Email, transaction.ID.Hex())
			}
		}
	}
}
