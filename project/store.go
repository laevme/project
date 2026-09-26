package main

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
)

type Store struct {
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

// isUniqueViolation ловит ошибку UNIQUE-констрейнта (код 23505)
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

// isForeignKeyViolation ловит нарушение внешнего ключа (код 23503)
func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23503"
	}
	return false
}

// ================= USERS =================

func (s *Store) ListUsers(ctx context.Context) ([]User, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, name, email, phone, created_at FROM users ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []User{}
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Phone, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (s *Store) GetUser(ctx context.Context, id int64) (User, error) {
	var u User
	err := s.db.QueryRow(ctx,
		`SELECT id, name, email, phone, created_at FROM users WHERE id = $1`, id).
		Scan(&u.ID, &u.Name, &u.Email, &u.Phone, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, err
	}
	return u, nil
}

func (s *Store) CreateUser(ctx context.Context, name, email string, phone *string) (User, error) {
	var u User
	err := s.db.QueryRow(ctx,
		`INSERT INTO users (name, email, phone) VALUES ($1, $2, $3)
		 RETURNING id, name, email, phone, created_at`,
		name, email, phone).
		Scan(&u.ID, &u.Name, &u.Email, &u.Phone, &u.CreatedAt)
	if isUniqueViolation(err) {
		return User{}, ErrConflict
	}
	if err != nil {
		return User{}, err
	}
	return u, nil
}

func (s *Store) UpdateUser(ctx context.Context, id int64, name, email string, phone *string) (User, error) {
	var u User
	err := s.db.QueryRow(ctx,
		`UPDATE users SET name = $1, email = $2, phone = $3 WHERE id = $4
		 RETURNING id, name, email, phone, created_at`,
		name, email, phone, id).
		Scan(&u.ID, &u.Name, &u.Email, &u.Phone, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if isUniqueViolation(err) {
		return User{}, ErrConflict
	}
	if err != nil {
		return User{}, err
	}
	return u, nil
}

func (s *Store) DeleteUser(ctx context.Context, id int64) error {
	tag, err := s.db.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ================= CATEGORIES =================

func (s *Store) ListCategories(ctx context.Context) ([]Category, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, name, slug FROM categories ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := []Category{}
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, rows.Err()
}

func (s *Store) GetCategory(ctx context.Context, id int64) (Category, error) {
	var c Category
	err := s.db.QueryRow(ctx,
		`SELECT id, name, slug FROM categories WHERE id = $1`, id).
		Scan(&c.ID, &c.Name, &c.Slug)
	if errors.Is(err, pgx.ErrNoRows) {
		return Category{}, ErrNotFound
	}
	if err != nil {
		return Category{}, err
	}
	return c, nil
}

func (s *Store) CreateCategory(ctx context.Context, name, slug string) (Category, error) {
	var c Category
	err := s.db.QueryRow(ctx,
		`INSERT INTO categories (name, slug) VALUES ($1, $2)
		 RETURNING id, name, slug`, name, slug).
		Scan(&c.ID, &c.Name, &c.Slug)
	if isUniqueViolation(err) {
		return Category{}, ErrConflict
	}
	if err != nil {
		return Category{}, err
	}
	return c, nil
}

func (s *Store) UpdateCategory(ctx context.Context, id int64, name, slug string) (Category, error) {
	var c Category
	err := s.db.QueryRow(ctx,
		`UPDATE categories SET name = $1, slug = $2 WHERE id = $3
		 RETURNING id, name, slug`, name, slug, id).
		Scan(&c.ID, &c.Name, &c.Slug)
	if errors.Is(err, pgx.ErrNoRows) {
		return Category{}, ErrNotFound
	}
	if isUniqueViolation(err) {
		return Category{}, ErrConflict
	}
	if err != nil {
		return Category{}, err
	}
	return c, nil
}

func (s *Store) DeleteCategory(ctx context.Context, id int64) error {
	tag, err := s.db.Exec(ctx, `DELETE FROM categories WHERE id = $1`, id)
	if isForeignKeyViolation(err) {
		return ErrConflict
	}
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ================= LISTINGS =================

func (s *Store) ListListings(ctx context.Context) ([]Listing, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, title, description, price, status, user_id, category_id, created_at
		 FROM listings ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	listings := []Listing{}
	for rows.Next() {
		var l Listing
		if err := rows.Scan(&l.ID, &l.Title, &l.Description, &l.Price,
			&l.Status, &l.UserID, &l.CategoryID, &l.CreatedAt); err != nil {
			return nil, err
		}
		listings = append(listings, l)
	}
	return listings, rows.Err()
}

func (s *Store) GetListing(ctx context.Context, id int64) (Listing, error) {
	var l Listing
	err := s.db.QueryRow(ctx,
		`SELECT id, title, description, price, status, user_id, category_id, created_at
		 FROM listings WHERE id = $1`, id).
		Scan(&l.ID, &l.Title, &l.Description, &l.Price,
			&l.Status, &l.UserID, &l.CategoryID, &l.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Listing{}, ErrNotFound
	}
	if err != nil {
		return Listing{}, err
	}
	return l, nil
}

func (s *Store) CreateListing(ctx context.Context, l Listing) (Listing, error) {
	err := s.db.QueryRow(ctx,
		`INSERT INTO listings (title, description, price, user_id, category_id)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, status, created_at`,
		l.Title, l.Description, l.Price, l.UserID, l.CategoryID).
		Scan(&l.ID, &l.Status, &l.CreatedAt)
	if isForeignKeyViolation(err) {
		return Listing{}, ErrNotFound
	}
	if err != nil {
		return Listing{}, err
	}
	return l, nil
}

func (s *Store) UpdateListing(ctx context.Context, id int64, l Listing) (Listing, error) {
	err := s.db.QueryRow(ctx,
		`UPDATE listings
		 SET title = $1, description = $2, price = $3, status = $4,
		     user_id = $5, category_id = $6
		 WHERE id = $7
		 RETURNING id, title, description, price, status, user_id, category_id, created_at`,
		l.Title, l.Description, l.Price, l.Status, l.UserID, l.CategoryID, id).
		Scan(&l.ID, &l.Title, &l.Description, &l.Price,
			&l.Status, &l.UserID, &l.CategoryID, &l.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Listing{}, ErrNotFound
	}
	if isForeignKeyViolation(err) {
		return Listing{}, ErrNotFound
	}
	if err != nil {
		return Listing{}, err
	}
	return l, nil
}

func (s *Store) DeleteListing(ctx context.Context, id int64) error {
	tag, err := s.db.Exec(ctx, `DELETE FROM listings WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ================= MESSAGES =================

func (s *Store) ListMessages(ctx context.Context) ([]Message, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, body, is_read, listing_id, sender_id, created_at
		 FROM messages ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := []Message{}
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.Body, &m.IsRead, &m.ListingID, &m.SenderID, &m.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, rows.Err()
}

func (s *Store) GetMessage(ctx context.Context, id int64) (Message, error) {
	var m Message
	err := s.db.QueryRow(ctx,
		`SELECT id, body, is_read, listing_id, sender_id, created_at
		 FROM messages WHERE id = $1`, id).
		Scan(&m.ID, &m.Body, &m.IsRead, &m.ListingID, &m.SenderID, &m.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Message{}, ErrNotFound
	}
	if err != nil {
		return Message{}, err
	}
	return m, nil
}

func (s *Store) CreateMessage(ctx context.Context, body string, listingID, senderID int64) (Message, error) {
	m := Message{Body: body, ListingID: listingID, SenderID: senderID}
	err := s.db.QueryRow(ctx,
		`INSERT INTO messages (body, listing_id, sender_id) VALUES ($1, $2, $3)
		 RETURNING id, is_read, created_at`, body, listingID, senderID).
		Scan(&m.ID, &m.IsRead, &m.CreatedAt)
	if isForeignKeyViolation(err) {
		return Message{}, ErrNotFound
	}
	if err != nil {
		return Message{}, err
	}
	return m, nil
}

func (s *Store) UpdateMessage(ctx context.Context, id int64, body string, isRead bool) (Message, error) {
	var m Message
	err := s.db.QueryRow(ctx,
		`UPDATE messages SET body = $1, is_read = $2 WHERE id = $3
		 RETURNING id, body, is_read, listing_id, sender_id, created_at`,
		body, isRead, id).
		Scan(&m.ID, &m.Body, &m.IsRead, &m.ListingID, &m.SenderID, &m.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Message{}, ErrNotFound
	}
	if err != nil {
		return Message{}, err
	}
	return m, nil
}

func (s *Store) DeleteMessage(ctx context.Context, id int64) error {
	tag, err := s.db.Exec(ctx, `DELETE FROM messages WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ================= FAVORITES =================

func (s *Store) ListFavorites(ctx context.Context) ([]Favorite, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, user_id, listing_id, created_at FROM favorites ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	favorites := []Favorite{}
	for rows.Next() {
		var f Favorite
		if err := rows.Scan(&f.ID, &f.UserID, &f.ListingID, &f.CreatedAt); err != nil {
			return nil, err
		}
		favorites = append(favorites, f)
	}
	return favorites, rows.Err()
}

func (s *Store) GetFavorite(ctx context.Context, id int64) (Favorite, error) {
	var f Favorite
	err := s.db.QueryRow(ctx,
		`SELECT id, user_id, listing_id, created_at FROM favorites WHERE id = $1`, id).
		Scan(&f.ID, &f.UserID, &f.ListingID, &f.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Favorite{}, ErrNotFound
	}
	if err != nil {
		return Favorite{}, err
	}
	return f, nil
}

func (s *Store) CreateFavorite(ctx context.Context, userID, listingID int64) (Favorite, error) {
	f := Favorite{UserID: userID, ListingID: listingID}
	err := s.db.QueryRow(ctx,
		`INSERT INTO favorites (user_id, listing_id) VALUES ($1, $2)
		 RETURNING id, created_at`, userID, listingID).
		Scan(&f.ID, &f.CreatedAt)
	if isUniqueViolation(err) {
		return Favorite{}, ErrConflict
	}
	if isForeignKeyViolation(err) {
		return Favorite{}, ErrNotFound
	}
	if err != nil {
		return Favorite{}, err
	}
	return f, nil
}

func (s *Store) UpdateFavorite(ctx context.Context, id int64, userID, listingID int64) (Favorite, error) {
	var f Favorite
	err := s.db.QueryRow(ctx,
		`UPDATE favorites SET user_id = $1, listing_id = $2 WHERE id = $3
		 RETURNING id, user_id, listing_id, created_at`,
		userID, listingID, id).
		Scan(&f.ID, &f.UserID, &f.ListingID, &f.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Favorite{}, ErrNotFound
	}
	if isUniqueViolation(err) {
		return Favorite{}, ErrConflict
	}
	if isForeignKeyViolation(err) {
		return Favorite{}, ErrNotFound
	}
	if err != nil {
		return Favorite{}, err
	}
	return f, nil
}

func (s *Store) DeleteFavorite(ctx context.Context, id int64) error {
	tag, err := s.db.Exec(ctx, `DELETE FROM favorites WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
