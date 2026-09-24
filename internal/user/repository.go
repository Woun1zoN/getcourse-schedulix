package user

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) GetByTelegramID(ctx context.Context, telegramID int64) (*User, error) {
	user := &User{}

	err := r.db.QueryRow(
		ctx,
		`SELECT id, telegram_id, getcourse_cookie, created_at
		FROM users
		WHERE telegram_id = $1`,
		telegramID,
	).Scan(&user.ID, &user.TelegramID, &user.GetCourseCookie, &user.CreatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return user, nil
}

func (r *Repository) Create(ctx context.Context, telegramID int64) (*User, error) {
	user := &User{}

	err := r.db.QueryRow(
		ctx,
		`INSERT INTO users (telegram_id)
		VALUES ($1)
		RETURNING id, telegram_id, getcourse_cookie, created_at`,
		telegramID,
	).Scan(&user.ID, &user.TelegramID, &user.GetCourseCookie, &user.CreatedAt)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *Repository) UpdateCookie(ctx context.Context, telegramID int64, cookie string,) error {
	_, err := r.db.Exec(
		ctx,
		`UPDATE users SET getcourse_cookie = $1 WHERE telegram_id = $2`,
		cookie,
		telegramID,
	)

	return err
}

func (r *Repository) GetActive(ctx context.Context) ([]User, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, telegram_id, getcourse_cookie, created_at
		FROM users
		WHERE getcourse_cookie IS NOT NULL`)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User

	for rows.Next() {
		var u User

		if err := rows.Scan(&u.ID, &u.TelegramID, &u.GetCourseCookie, &u.CreatedAt); err != nil {
			return nil, err
		}

		users = append(users, u)
	}

	return users, rows.Err()
}