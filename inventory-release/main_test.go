package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"e-commerce-app/models"
	"e-commerce-app/utils"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

const (
    testOrder = `
    {
        "order_id": "orderID123456", 
        "order_info": 
        {
            "order_date": "2022-01-01T02:30:50Z",
            "customer_id": "id001",
            "order_status": "processing",
            "items": [
                {
                "item_id": "itemID456",
                "qty": 1,
                "description": "Pencil",
                "unit_price": 2.5
                },
                {
                "item_id": "itemID789",
                "qty": 1,
                "description": "Paper",
                "unit_price": 4.0
                }
            ],
            "payment": {
                "merchant_id": "merchantID1234",
                "payment_amount": 6.5,
                "transaction_id": "transactionID7845764",
                "transaction_date": "01-1-2022",
                "order_id": "orderID123456",
                "payment_type": "creditcard"
            },
            "inventory": {
                "transaction_id": "transactionID7845764",
                "transaction_date": "01-1-2022",
                "order_id": "orderID123456",
                "items": ["itemID456", "itemID789"],
                "transaction_type": "online"
            }
        }
    }`
)

func TestHandler(t *testing.T) {
    assert := assert.New(t)

    t.Run("ReleaseInventory", func(t *testing.T) {

        prepareDb(t)
        createTable(t)
        stoOrd := prepareTestData(t)

        req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(testOrder))
        resp := httptest.NewRecorder()

        handler(resp, req)

        // Check database
        test(t, assert, stoOrd)

        // Clean up
        cleanUp(t)
    })
}

func prepareDb(t *testing.T) {
    utils.CredsLocation = "../test/postgres-creds.json"
    utils.SSLMode = "disable"

    var err error
    db, err = utils.ConnectDatabase()
    if err != nil {
        t.Fatalf("error connecting to the db: %v", err)
    }
}

func createTable(t *testing.T) {
    _, err := db.Exec(`CREATE TABLE stored_orders (order_id text, order_info JSONB);`)
    if err, ok := err.(*pq.Error); ok && err.Code.Name() != "duplicate_table" {
        t.Fatalf("createTable(): error creating table %v", err)
    }
}

func prepareTestData(t *testing.T) models.StoredOrder {
    // Parsing test data prior to calling handler()
    stoOrd := models.StoredOrder{}
    err := json.Unmarshal([]byte(testOrder), &stoOrd)
    if err != nil {
        t.Fatalf("prepareTestData(): error with json unmarshall: %v", err)
    }

    insertCommand := `INSERT INTO stored_orders (order_id, order_info) VALUES ($1, $2)`
    _, err = db.Exec(insertCommand, stoOrd.OrderID, stoOrd.Order)
    if err != nil {
        fmt.Printf("prepareTestData(): Could not insert test data into database: %v", err)
    }

    return stoOrd
}

func test(t *testing.T, a *assert.Assertions, stoOrd models.StoredOrder) {
    // Fetching test data from test database after calling handler()
    var allOrderInfos []models.StoredOrder
    var storedOrder models.StoredOrder
    rows, err := db.Query(`SELECT * FROM stored_orders WHERE order_id=$1`, stoOrd.OrderID)
    if err != nil {
        t.Fatalf("test(): error with query: %v", err)
    }

    // Parsing data from database
    for rows.Next() {
        if err = rows.Scan(&storedOrder.OrderID, &storedOrder.Order); err != nil {
            if err != nil {
                t.Fatalf("test(): error with scan: %v", err)
            }
        } 
        // Scan worked, so run asserts        
        a.True(storedOrder.Order.Inventory.TransactionID != stoOrd.Order.Inventory.TransactionID, "TransactionID should be modified")
        a.True(storedOrder.Order.Inventory.TransactionDate != stoOrd.Order.Inventory.TransactionDate, "TransactionDate should be modified")
        a.True(storedOrder.Order.Inventory.TransactionType == "Release", "Transaction type should be equal to 'Release'")
        allOrderInfos = append(allOrderInfos, storedOrder)
    }
    fmt.Println("Printing all orders:\n", allOrderInfos)
}

func cleanUp(t *testing.T) {
    // Cleanup table
    _, err := db.Exec(`TRUNCATE stored_orders;`)
    if err != nil {
        t.Fatalf("cleanUp(): error truncating data in table %v", err)
    }
    _, err = db.Exec(`DELETE FROM stored_orders;`)
    if err != nil {
        t.Fatalf("cleanUp(): error deleting data in table %v", err)
    }
}