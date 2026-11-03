package user

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const userColumns = `id, telegram_id, getcourse_cookie, created_at, getcourse_stream_ids, target_chat_id, target_chat_title`

type rowScanner interface {
	Scan(dest ...any) error
}

func scanUser(row rowScanner, u *User) error {
	return row.Scan(
		&u.ID,
		&u.TelegramID,
		&u.GetCourseCookie,
		&u.CreatedAt,
		&u.GetCourseStreamIDs,
		&u.TargetChatID,
		&u.TargetChatTitle,
	)
}

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
	err := scanUser(r.db.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE telegram_id = $1`, telegramID), user)
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

	err := scanUser(r.db.QueryRow(ctx,
		`INSERT INTO users (telegram_id) VALUES ($1) RETURNING `+userColumns, telegramID), user)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *Repository) UpdateCookie(ctx context.Context, telegramID int64, cookie string) error {
	_, err := r.db.Exec(
		ctx,
		`UPDATE users SET getcourse_cookie = $1 WHERE telegram_id = $2`,
		cookie,
		telegramID,
	)

	return err
}

func (r *Repository) GetActive(ctx context.Context) ([]User, error) {
	rows, err := r.db.Query(ctx, `SELECT `+userColumns+` FROM users WHERE getcourse_cookie IS NOT NULL`)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User

	for rows.Next() {
		var u User

		if err := scanUser(rows, &u); err != nil {
			return nil, err
		}

		users = append(users, u)
	}

	return users, rows.Err()
}

func (r *Repository) UpdateStreams(ctx context.Context, telegramID int64, streamIDs []int64) error {
	if streamIDs == nil {
		streamIDs = []int64{}
	}

	_, err := r.db.Exec(ctx,
		`UPDATE users SET getcourse_stream_ids = $1 WHERE telegram_id = $2`,
		streamIDs, telegramID)
	return err
}

func (r *Repository) UpdateTargetChat(ctx context.Context, telegramID int64, chatID int64, title string) error {
	_, err := r.db.Exec(
		ctx,
		`UPDATE users SET target_chat_id = $1, target_chat_title = $2 WHERE telegram_id = $3`,
		chatID,
		title,
		telegramID,
	)

	return err
}

func (r *Repository) ResetTargetChat(ctx context.Context, telegramID int64) error {
	_, err := r.db.Exec(ctx,
		`UPDATE users SET target_chat_id = telegram_id, target_chat_title = NULL WHERE telegram_id = $1`,
		telegramID)

	return err
}

func (r *Repository) ClearTargetChat(ctx context.Context, telegramID int64) error {
	_, err := r.db.Exec(ctx,
		`UPDATE users SET target_chat_id = NULL, target_chat_title = NULL WHERE telegram_id = $1`,
		telegramID)

	return err
}

func (r *Repository) ClearStream(ctx context.Context, telegramID int64) error {
	_, err := r.db.Exec(ctx,
		`UPDATE users SET getcourse_stream_ids = '{}' WHERE telegram_id = $1`, telegramID)
	return err
}

func (r *Repository) DeleteByTelegramID(ctx context.Context, telegramID int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM users WHERE telegram_id = $1`, telegramID)
	return err
}
