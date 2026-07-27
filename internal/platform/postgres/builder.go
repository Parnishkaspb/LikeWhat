// Package postgres contains PostgreSQL-specific infrastructure shared by repositories.
package postgres

import squirrel "github.com/Masterminds/squirrel"

// Builder generates parameterized SQL queries compatible with PostgreSQL.
// Repositories must use it instead of concatenating request data into SQL.
var Builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
