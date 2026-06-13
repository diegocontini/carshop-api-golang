package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"utfpr.edu.br/carshop-api/src/domain"
	"utfpr.edu.br/carshop-api/src/infra/sqlc"
)

type UserService struct {
	pool *pgxpool.Pool
	q    *sqlc.Queries
}

func NewUserService(pool *pgxpool.Pool) *UserService {
	return &UserService{pool: pool, q: sqlc.New(pool)}
}

func (s *UserService) List(ctx context.Context) ([]domain.User, error) {
	rows, err := s.q.ListUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	out := make([]domain.User, len(rows))
	for i, r := range rows {
		out[i] = userFromSQLC(r)
	}
	return out, nil
}

func (s *UserService) Get(ctx context.Context, id int64) (domain.User, error) {
	row, err := s.q.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.ErrNotFound
		}
		return domain.User{}, fmt.Errorf("get user: %w", err)
	}
	return userFromSQLC(row), nil
}

func (s *UserService) Create(ctx context.Context, in domain.User) (domain.User, error) {
	if !in.Role.IsValid() {
		return domain.User{}, fmt.Errorf("%w: role must be admin or vendor", domain.ErrInvalid)
	}
	hash, err := HashPassword(in.Password)
	if err != nil {
		return domain.User{}, fmt.Errorf("hash password: %w", err)
	}
	row, err := s.q.CreateUser(ctx, sqlc.CreateUserParams{
		Username:                  in.Username,
		Password:                  hash,
		Email:                     in.Email,
		ComissionPerSaleInPercent: in.ComissionPerSaleInPercent,
		Role:                      string(in.Role),
	})
	if err != nil {
		if isUniqueViolation(err) {
			return domain.User{}, fmt.Errorf("%w: username already exists", domain.ErrConflict)
		}
		return domain.User{}, fmt.Errorf("create user: %w", err)
	}
	return userFromSQLC(row), nil
}

func (s *UserService) Update(ctx context.Context, id int64, in domain.User) (domain.User, error) {
	if !in.Role.IsValid() {
		return domain.User{}, fmt.Errorf("%w: role must be admin or vendor", domain.ErrInvalid)
	}
	hash, err := HashPassword(in.Password)
	if err != nil {
		return domain.User{}, fmt.Errorf("hash password: %w", err)
	}
	row, err := s.q.UpdateUser(ctx, sqlc.UpdateUserParams{
		ID:                        id,
		Username:                  in.Username,
		Password:                  hash,
		Email:                     in.Email,
		ComissionPerSaleInPercent: in.ComissionPerSaleInPercent,
		Role:                      string(in.Role),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.ErrNotFound
		}
		if isUniqueViolation(err) {
			return domain.User{}, fmt.Errorf("%w: username already exists", domain.ErrConflict)
		}
		return domain.User{}, fmt.Errorf("update user: %w", err)
	}
	return userFromSQLC(row), nil
}

func (s *UserService) Delete(ctx context.Context, id int64) error {
	n, err := s.q.DeleteUser(ctx, id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// Authenticate returns the user if credentials match (bcrypt). Returns
// ErrNotFound for both unknown usernames and wrong passwords so the
// caller cannot use the error to distinguish (avoids username enumeration).
func (s *UserService) Authenticate(ctx context.Context, username, password string) (domain.User, error) {
	row, err := s.q.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.ErrNotFound
		}
		return domain.User{}, fmt.Errorf("lookup user: %w", err)
	}
	if !CheckPassword(row.Password, password) {
		return domain.User{}, domain.ErrNotFound
	}
	return userFromSQLC(row), nil
}

// SeedAdmin creates the configured superuser if missing. Idempotent across
// concurrent boots via the SeedAdmin query's ON CONFLICT clause.
func (s *UserService) SeedAdmin(ctx context.Context, username, plaintextPassword, email string) error {
	hash, err := HashPassword(plaintextPassword)
	if err != nil {
		return fmt.Errorf("hash admin password: %w", err)
	}
	percent := int16(3)
	return s.q.SeedAdmin(ctx, sqlc.SeedAdminParams{
		Username:                  username,
		Password:                  hash,
		Email:                     email,
		ComissionPerSaleInPercent: &percent,
	})
}

func userFromSQLC(u sqlc.User) domain.User {
	return domain.User{
		ID:                        u.ID,
		Username:                  u.Username,
		Password:                  u.Password,
		Email:                     u.Email,
		ComissionPerSaleInPercent: u.ComissionPerSaleInPercent,
		Role:                      domain.UserRole(u.Role),
	}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
