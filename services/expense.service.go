package services

import (
	"errors"
	"math"

	"github.com/ebubekiryigit/golang-mongodb-rest-api-starter/models"
	db "github.com/ebubekiryigit/golang-mongodb-rest-api-starter/models/db"
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateExpense(userID primitive.ObjectID, req models.CreateExpenseRequest) (*db.Expense, error) {
	groupID, err := primitive.ObjectIDFromHex(req.GroupID)
	if err != nil {
		return nil, errors.New("invalid group ID")
	}

	// Check if user is group member
	_, err = GetGroupById(groupID, userID)
	if err != nil {
		return nil, err
	}

	splitType := db.SplitType(req.SplitType)
	expense := db.NewExpense(groupID, req.Description, req.Amount, req.Currency, userID, splitType, req.Category)
	expense.Notes = req.Notes

	// Process splits
	if len(req.Splits) == 0 {
		return nil, errors.New("at least one split is required")
	}

	var totalSplit float64
	for _, split := range req.Splits {
		splitUserID, err := primitive.ObjectIDFromHex(split.UserID)
		if err != nil {
			return nil, errors.New("invalid user ID in split")
		}

		expense.Splits = append(expense.Splits, db.ExpenseSplit{
			UserID: splitUserID,
			Amount: split.Amount,
		})
		totalSplit += split.Amount
	}

	// Validate splits based on type
	switch splitType {
	case db.SplitTypeEqual:
		// Auto-calculate equal splits
		splitAmount := req.Amount / float64(len(req.Splits))
		for i := range expense.Splits {
			expense.Splits[i].Amount = math.Round(splitAmount*100) / 100 // Round to 2 decimals
		}
	case db.SplitTypeExact:
		// Check if splits add up to total amount
		if math.Abs(totalSplit-req.Amount) > 0.01 {
			return nil, errors.New("split amounts must add up to total expense amount")
		}
	case db.SplitTypePercentage:
		// Check if percentages add up to 100
		if math.Abs(totalSplit-100.0) > 0.01 {
			return nil, errors.New("split percentages must add up to 100")
		}
		// Convert percentages to amounts
		for i := range expense.Splits {
			expense.Splits[i].Amount = (expense.Splits[i].Amount / 100.0) * req.Amount
		}
	}

	err = mgm.Coll(expense).Create(expense)
	if err != nil {
		return nil, err
	}

	return expense, nil
}

func GetGroupExpenses(groupID, userID primitive.ObjectID, page, limit int) ([]*db.Expense, error) {
	// Check if user is group member
	_, err := GetGroupById(groupID, userID)
	if err != nil {
		return nil, err
	}

	var expenses []*db.Expense
	findOptions := options.Find().
		SetSkip(int64(page * limit)).
		SetLimit(int64(limit + 1)).   // +1 to check if there are more
		SetSort(bson.D{{"date", -1}}) // Latest first

	err = mgm.Coll(&db.Expense{}).SimpleFind(&expenses, bson.M{
		"group_id": groupID,
	}, findOptions)

	return expenses, err
}

func GetExpenseById(expenseID, userID primitive.ObjectID) (*db.Expense, error) {
	expense := &db.Expense{}

	err := mgm.Coll(expense).FindByID(expenseID, expense)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("expense not found")
		}
		return nil, err
	}

	// Check if user is group member
	_, err = GetGroupById(expense.GroupID, userID)
	if err != nil {
		return nil, err
	}

	return expense, nil
}

func UpdateExpense(expenseID, userID primitive.ObjectID, req models.UpdateExpenseRequest) error {
	expense, err := GetExpenseById(expenseID, userID)
	if err != nil {
		return err
	}

	// Only the person who paid can update the expense
	if expense.PaidBy != userID {
		return errors.New("only the payer can update this expense")
	}

	updateDoc := bson.M{}

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

	// Handle split updates
	if req.SplitType != "" && len(req.Splits) > 0 {
		splitType := db.SplitType(req.SplitType)
		var splits []db.ExpenseSplit

		amount := expense.Amount
		if req.Amount > 0 {
			amount = req.Amount
		}

		var totalSplit float64
		for _, split := range req.Splits {
			splitUserID, err := primitive.ObjectIDFromHex(split.UserID)
			if err != nil {
				return errors.New("invalid user ID in split")
			}

			splits = append(splits, db.ExpenseSplit{
				UserID: splitUserID,
				Amount: split.Amount,
			})
			totalSplit += split.Amount
		}

		// Validate splits based on type
		switch splitType {
		case db.SplitTypeEqual:
			splitAmount := amount / float64(len(splits))
			for i := range splits {
				splits[i].Amount = math.Round(splitAmount*100) / 100
			}
		case db.SplitTypeExact:
			if math.Abs(totalSplit-amount) > 0.01 {
				return errors.New("split amounts must add up to total expense amount")
			}
		case db.SplitTypePercentage:
			if math.Abs(totalSplit-100.0) > 0.01 {
				return errors.New("split percentages must add up to 100")
			}
			for i := range splits {
				splits[i].Amount = (splits[i].Amount / 100.0) * amount
			}
		}

		updateDoc["split_type"] = splitType
		updateDoc["splits"] = splits
	}

	if len(updateDoc) == 0 {
		return errors.New("no fields to update")
	}

	_, err = mgm.Coll(expense).UpdateOne(mgm.Ctx(), bson.M{"_id": expenseID}, bson.M{
		"$set": updateDoc,
	})

	return err
}

func DeleteExpense(expenseID, userID primitive.ObjectID) error {
	expense, err := GetExpenseById(expenseID, userID)
	if err != nil {
		return err
	}

	// Only the person who paid can delete the expense
	if expense.PaidBy != userID {
		return errors.New("only the payer can delete this expense")
	}

	err = mgm.Coll(expense).Delete(expense)
	return err
}

func GetUserExpenses(userID primitive.ObjectID, page, limit int) ([]*db.Expense, error) {
	var expenses []*db.Expense

	// Find expenses where user is either the payer or in the splits
	findOptions := options.Find().
		SetSkip(int64(page * limit)).
		SetLimit(int64(limit + 1)).
		SetSort(bson.D{{"date", -1}})

	err := mgm.Coll(&db.Expense{}).SimpleFind(&expenses, bson.M{
		"$or": []bson.M{
			{"paid_by": userID},
			{"splits.user_id": userID},
		},
	}, findOptions)

	return expenses, err
}
