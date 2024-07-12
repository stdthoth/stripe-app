package models

import (
	"context"
	"database/sql"
	"time"
)

type DBmodels struct {
	DB *sql.DB
}

type Models struct {
	DB DBmodels
}

// GopherImages is our product structure
type GopherImages struct {
	Id             int       `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	InventoryLevel int       `json:"inventory_level"`
	Image          string    `json:"image"`
	Price          int       `json:"price"`
	IsRecurring    bool      `json:"is_recurring"`
	PlanID         string    `json:"plan_id"`
	CreatedAt      time.Time `json:"-"`
	UpdatedAt      time.Time `json:"-"`
}

type Order struct {
	ID            int       `json:"id"`
	ImageID       int       `json:"image_id"`
	TransactionID int       `json:"transaction_id"`
	CustomerID    int       `json:"customer_id"`
	StatusID      int       `json:"status_id"`
	Quantity      int       `json:"quantity"`
	Amount        int       `json:"amount"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Status struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TransactionStatus struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Transaction struct {
	ID                  int       `json:"id"`
	Amount              int       `json:"amount"`
	Currency            string    `json:"currency"`
	LastFour            string    `json:"last_four"`
	BankReturnCode      string    `json:"bank_return_code"`
	ExpiryMonth         int       `json:"expiry_month"`
	ExpiryYear          int       `json:"expiry_year"`
	PaymentIntent       string    `json:"payment_intent"`
	PaymentMethod       string    `json:"payment_method"`
	TransactionStatusID int       `json:"transaction_status_id"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type User struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Customer struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewModels(db *sql.DB) Models {
	return Models{
		DB: DBmodels{
			DB: db,
		},
	}
}

func (m *DBmodels) GetGopherImages(id int) (GopherImages, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	var image GopherImages

	row := m.DB.QueryRowContext(ctx, `select 
		id, name, description, inventory_level, price, coalesce(image, ''),
		is_recurring, plan_id, 
		created_at, updated_at
	 from 
	 	GopherImages 
	 where id = ?`, id)
	err := row.Scan(
		&image.Id,
		&image.Name,
		&image.Description,
		&image.InventoryLevel,
		&image.Price,
		&image.Image,
		&image.IsRecurring,
		&image.PlanID,
		&image.CreatedAt,
		&image.UpdatedAt)
	if err != nil {
		return image, err
	}

	return image, nil
}

func (m *DBmodels) InsertTransaction(tx Transaction) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	sqlstmt := `
		insert into transactions
			(amount, currency, last_four, bank_return_code, expiry_month, expiry_year
			payment_intent,payment_method, transaction_status_id, created_at, updated_at)
		values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	
	`
	result, err := m.DB.ExecContext(ctx, sqlstmt,
		tx.Amount,
		tx.Currency,
		tx.LastFour,
		tx.BankReturnCode,
		tx.ExpiryMonth,
		tx.ExpiryYear,
		tx.TransactionStatusID,
		tx.PaymentIntent,
		tx.PaymentMethod,
		time.Now(),
		time.Now())

	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), err
}

func (m *DBmodels) InsertOrder(ord Order) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	sqlstmt := `
		insert into transactions
			(image_id, transaction_id, status_id, quantity, customer_id, 
			amount, created_at, updated_at)
		values(?, ?, ?, ?, ?, ?, ?, ?)
	
	`
	result, err := m.DB.ExecContext(ctx, sqlstmt,
		ord.ImageID,
		ord.TransactionID,
		ord.StatusID,
		ord.Quantity,
		ord.CustomerID,
		ord.Amount,
		time.Now(),
		time.Now())

	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), err
}

func (m *DBmodels) InsertCustomer(customer Customer) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	sqlstmt := `
		insert into transactions
			(first_name, last_name, email, created_at, updated_at)
		values(?, ?, ?, ?, ?)
	
	`
	result, err := m.DB.ExecContext(ctx, sqlstmt,
		customer.FirstName,
		customer.LastName,
		customer.Email,
		time.Now(),
		time.Now())

	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), err
}
