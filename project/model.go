package main

import "time"

type User struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     *string   `json:"phone"`
	CreatedAt time.Time `json:"created_at"`
}

type Category struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type Listing struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Status      string    `json:"status"`
	UserID      int64     `json:"user_id"`
	CategoryID  int64     `json:"category_id"`
	CreatedAt   time.Time `json:"created_at"`
}

type Message struct {
	ID        int64     `json:"id"`
	Body      string    `json:"body"`
	IsRead    bool      `json:"is_read"`
	ListingID int64     `json:"listing_id"`
	SenderID  int64     `json:"sender_id"`
	CreatedAt time.Time `json:"created_at"`
}

type Favorite struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	ListingID int64     `json:"listing_id"`
	CreatedAt time.Time `json:"created_at"`
}
