package models

// Table and column names of the users table. They live next to the domain
// model so repositories share a single source of truth for SQL fragments.
const (
	TableUsers = "users"

	ColUserID         = "id"
	ColUserTelegramID = "telegram_id"
	ColUserNickName   = "nick_name"
	ColUserName       = "name"
)

// User is the domain model of a chat user. It deliberately does not depend on
// protobuf. Per the spec it carries no created/updated/deleted timestamps.
type User struct {
	ID         int64
	TelegramID int64
	NickName   string
	Name       string
}

// UserFilter limits user records returned by a repository.
type UserFilter struct {
	NameLike string
}
