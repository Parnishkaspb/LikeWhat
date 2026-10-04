package user

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/db"
	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/models"
	pg "github.com/Parnishkaspb/LikeWhat/internal/platform/postgres"
)

// Store holds user queries and shares the connection pool with the other
// entity stores.
type Store struct {
	*db.Client
}

// NewStore creates a user store on top of the shared client.
func NewStore(client *db.Client) *Store {
	return &Store{Client: client}
}

func (s *Store) CreateUser(ctx context.Context, item models.User) (models.User, error) {
	query, args, err := pg.Builder.
		Insert(models.TableUsers).
		Columns(models.ColUserTelegramID, models.ColUserNickName, models.ColUserName).
		Values(item.TelegramID, item.NickName, item.Name).
		Suffix("RETURNING " + models.ColUserID).
		ToSql()
	if err != nil {
		return models.User{}, err
	}

	if err := s.QueryRow(ctx, query, args...).Scan(&item.ID); err != nil {
		return models.User{}, err
	}
	return item, nil
}

func (s *Store) GetUser(ctx context.Context, id int64) (models.User, error) {
	query, args, err := pg.Builder.
		Select(models.ColUserID, models.ColUserTelegramID, models.ColUserNickName, models.ColUserName).
		From(models.TableUsers).
		Where(squirrel.Eq{models.ColUserID: id}).
		ToSql()
	if err != nil {
		return models.User{}, err
	}

	var item models.User
	if err := s.QueryRow(ctx, query, args...).
		Scan(&item.ID, &item.TelegramID, &item.NickName, &item.Name); err != nil {
		return models.User{}, db.NotFound(err)
	}
	return item, nil
}

func (s *Store) ListUsers(ctx context.Context, filter models.UserFilter) ([]models.User, error) {
	builder := pg.Builder.
		Select(models.ColUserID, models.ColUserTelegramID, models.ColUserNickName, models.ColUserName).
		From(models.TableUsers).
		OrderBy(models.ColUserID)

	if filter.NameLike != "" {
		builder = builder.Where(db.ILikeExpr(models.ColUserName, filter.NameLike))
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := s.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.User
	for rows.Next() {
		var item models.User
		if err := rows.Scan(&item.ID, &item.TelegramID, &item.NickName, &item.Name); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
