package utils

import (
	"database/sql"
	"e-commerce-app/models"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func CheckForErrors(err error, s string) {
	if err != nil {
		fmt.Printf("%v\n", err)
		log.Fatalf(s)
	}
}

func ConnectDatabase() (*sql.DB, error) {
	// connection string
	host := "localhost"
    port := 5432
    user := "mruizcardenas"
    password := "K67u5ye"
    dbname := "postgres"

	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	
	// open database
	db, err := sql.Open("postgres", psqlconn)
	CheckForErrors(err, "Could not open database")

	// check db
    err = db.Ping()
	CheckForErrors(err, "Could not ping database")
	fmt.Println("Connected to databse!")
	return db, err
}

func ViewDatabase(db *sql.DB) {
	var allStoredOrders []models.StoredOrder
	var storedOrder models.StoredOrder
	rows, err := db.Query(`SELECT * FROM stored_orders`)
	CheckForErrors(err, "send: Could not query select * from stored_orders")

	for rows.Next() {
		if err = rows.Scan(&storedOrder.OrderID, &storedOrder.Order); err != nil {
			CheckForErrors(err, "Error with scan")
		} else {
			// fmt.Println("Here's where scan has no error")
		}
		allStoredOrders = append(allStoredOrders, storedOrder)
	}

	fmt.Println(allStoredOrders)
}

func ResetDatabase(db *sql.DB) {
	// Resetting after inventory-reserve
	originalInventory := `UPDATE stored_orders SET order_info = jsonb_set(order_info, '{inventory}', '{
		"transaction_id": "transactionID7845764", 
		"transaction_date": "01-1-2022", 
		"order_id": "orderID123456", 
		"items": [
			"Pencil", 
			"Paper"
		], 
		"transaction_type": "online"
	}', true) WHERE order_id = 'orderID123456';`

	_, err := db.Exec(originalInventory)
	CheckForErrors(err, "Could not reset database")
}