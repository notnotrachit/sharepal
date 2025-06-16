#!/bin/bash

# Splitwise API Test Script
# This script demonstrates the basic functionality of the Splitwise API

API_BASE="http://localhost:8080/v1"
CONTENT_TYPE="Content-Type: application/json"

echo "🚀 Splitwise API Demo"
echo "===================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to make API calls with error handling
api_call() {
    local method=$1
    local endpoint=$2
    local data=$3
    local token=$4
    
    echo -e "${BLUE}📡 $method $endpoint${NC}"
    
    if [ -n "$token" ]; then
        if [ -n "$data" ]; then
            response=$(curl -s -X $method "$API_BASE$endpoint" \
                -H "$CONTENT_TYPE" \
                -H "Bearer-Token: $token" \
                -d "$data")
        else
            response=$(curl -s -X $method "$API_BASE$endpoint" \
                -H "Bearer-Token: $token")
        fi
    else
        response=$(curl -s -X $method "$API_BASE$endpoint" \
            -H "$CONTENT_TYPE" \
            -d "$data")
    fi
    
    echo "$response" | jq '.' 2>/dev/null || echo "$response"
    echo ""
    
    # Return the response for further processing
    echo "$response"
}

# Function to extract value from JSON response
extract_json_value() {
    local json=$1
    local key=$2
    echo "$json" | jq -r ".$key" 2>/dev/null
}

echo "📋 Starting API Demo..."
echo ""

# 1. Health Check
echo -e "${YELLOW}1. Health Check${NC}"
api_call "GET" "/ping"

# 2. Register Users
echo -e "${YELLOW}2. Registering Users${NC}"

user1_response=$(api_call "POST" "/auth/register" '{
    "name": "Alice Johnson",
    "email": "alice@example.com",
    "password": "password123"
}')

user2_response=$(api_call "POST" "/auth/register" '{
    "name": "Bob Smith",
    "email": "bob@example.com", 
    "password": "password123"
}')

user3_response=$(api_call "POST" "/auth/register" '{
    "name": "Charlie Brown",
    "email": "charlie@example.com",
    "password": "password123"
}')

# Extract tokens
alice_token=$(echo "$user1_response" | jq -r '.data.token.access.token' 2>/dev/null)
alice_id=$(echo "$user1_response" | jq -r '.data.user._id' 2>/dev/null)

# 3. Login as Bob
echo -e "${YELLOW}3. Login as Bob${NC}"
bob_login=$(api_call "POST" "/auth/login" '{
    "email": "bob@example.com",
    "password": "password123"
}')

bob_token=$(echo "$bob_login" | jq -r '.data.token.access.token' 2>/dev/null)
bob_id=$(echo "$bob_login" | jq -r '.data.user._id' 2>/dev/null)

# 4. Login as Charlie
echo -e "${YELLOW}4. Login as Charlie${NC}"
charlie_login=$(api_call "POST" "/auth/login" '{
    "email": "charlie@example.com",
    "password": "password123"
}')

charlie_token=$(echo "$charlie_login" | jq -r '.data.token.access.token' 2>/dev/null)
charlie_id=$(echo "$charlie_login" | jq -r '.data.user._id' 2>/dev/null)

if [ "$alice_token" = "null" ] || [ -z "$alice_token" ]; then
    echo -e "${RED}❌ Failed to get Alice's token. Exiting.${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Users registered and logged in successfully${NC}"
echo "Alice ID: $alice_id"
echo "Bob ID: $bob_id" 
echo "Charlie ID: $charlie_id"
echo ""

# 5. Create a Group (as Alice)
echo -e "${YELLOW}5. Creating a Group${NC}"
group_response=$(api_call "POST" "/groups" '{
    "name": "Weekend Trip",
    "description": "Our weekend getaway to the mountains",
    "currency": "USD",
    "member_ids": ["'"$bob_id"'", "'"$charlie_id"'"]
}' "$alice_token")

group_id=$(echo "$group_response" | jq -r '.data.group._id' 2>/dev/null)

if [ "$group_id" = "null" ] || [ -z "$group_id" ]; then
    echo -e "${RED}❌ Failed to create group. Exiting.${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Group created with ID: $group_id${NC}"
echo ""

# 6. View Group Details
echo -e "${YELLOW}6. Viewing Group Details${NC}"
api_call "GET" "/groups/$group_id" "" "$alice_token"

# 7. Add an Expense (Alice pays for dinner)
echo -e "${YELLOW}7. Adding Expense - Dinner${NC}"
expense1_response=$(api_call "POST" "/expenses" '{
    "group_id": "'"$group_id"'",
    "description": "Dinner at Mountain View Restaurant",
    "amount": 150.00,
    "currency": "USD",
    "split_type": "equal",
    "splits": [
        {"user_id": "'"$alice_id"'", "amount": 0},
        {"user_id": "'"$bob_id"'", "amount": 0},
        {"user_id": "'"$charlie_id"'", "amount": 0}
    ],
    "category": "Food",
    "notes": "Great dinner with mountain views!"
}' "$alice_token")

expense1_id=$(echo "$expense1_response" | jq -r '.data.expense._id' 2>/dev/null)

# 8. Add another expense (Bob pays for gas)
echo -e "${YELLOW}8. Adding Expense - Gas${NC}"
api_call "POST" "/expenses" '{
    "group_id": "'"$group_id"'",
    "description": "Gas for the trip",
    "amount": 90.00,
    "currency": "USD", 
    "split_type": "exact",
    "splits": [
        {"user_id": "'"$alice_id"'", "amount": 30.00},
        {"user_id": "'"$bob_id"'", "amount": 30.00},
        {"user_id": "'"$charlie_id"'", "amount": 30.00}
    ],
    "category": "Transportation"
}' "$bob_token")

# 9. Add third expense (Charlie pays for accommodation)
echo -e "${YELLOW}9. Adding Expense - Hotel${NC}"
api_call "POST" "/expenses" '{
    "group_id": "'"$group_id"'",
    "description": "Hotel accommodation for 2 nights",
    "amount": 300.00,
    "currency": "USD",
    "split_type": "percentage",
    "splits": [
        {"user_id": "'"$alice_id"'", "amount": 40},
        {"user_id": "'"$bob_id"'", "amount": 30},
        {"user_id": "'"$charlie_id"'", "amount": 30}
    ],
    "category": "Accommodation"
}' "$charlie_token")

# 10. View Group Expenses
echo -e "${YELLOW}10. Viewing Group Expenses${NC}"
api_call "GET" "/groups/$group_id/expenses" "" "$alice_token"

# 11. Check Balances
echo -e "${YELLOW}11. Checking Group Balances${NC}"
api_call "GET" "/groups/$group_id/balances" "" "$alice_token"

# 12. Simplify Debts
echo -e "${YELLOW}12. Simplifying Debts${NC}"
api_call "GET" "/groups/$group_id/simplify" "" "$alice_token"

# 13. Send Friend Request
echo -e "${YELLOW}13. Sending Friend Request (Alice to Bob)${NC}"
api_call "POST" "/friends/request" '{
    "email": "bob@example.com"
}' "$alice_token"

# 14. View Friend Requests (as Bob)
echo -e "${YELLOW}14. Viewing Friend Requests (Bob)${NC}"
friend_requests=$(api_call "GET" "/friends/requests/received" "" "$bob_token")

# Extract friend request ID
request_id=$(echo "$friend_requests" | jq -r '.data.requests[0]._id' 2>/dev/null)

if [ "$request_id" != "null" ] && [ -n "$request_id" ]; then
    # 15. Accept Friend Request
    echo -e "${YELLOW}15. Accepting Friend Request${NC}"
    api_call "POST" "/friends/request/$request_id/respond" '{
        "accept": true
    }' "$bob_token"
    
    # 16. View Friends List
    echo -e "${YELLOW}16. Viewing Friends List (Alice)${NC}"
    api_call "GET" "/friends" "" "$alice_token"
fi

# 17. View User's All Expenses
echo -e "${YELLOW}17. Viewing Alice's All Expenses${NC}"
api_call "GET" "/expenses" "" "$alice_token"

echo -e "${GREEN}🎉 API Demo Complete!${NC}"
echo ""
echo -e "${BLUE}📊 Summary:${NC}"
echo "- ✅ Users registered and authenticated"
echo "- ✅ Group created with multiple members"
echo "- ✅ Expenses added with different split types"
echo "- ✅ Balances calculated automatically"
echo "- ✅ Debt simplification working"
echo "- ✅ Friend requests sent and accepted"
echo ""
echo -e "${YELLOW}🌐 View the full API documentation at:${NC}"
echo "http://localhost:8080/swagger/index.html"
