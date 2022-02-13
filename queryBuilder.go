package queryBuilder

import (
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type QB struct {
	db *gorm.DB

	// support field for QueryBuilder interface
	// Used for tracing during building sql query.
	// Must be initialized separately for each query.
	logTrace logrus.Fields

	projectName string

	errInternal error
	errNotFound error
}

func New(db *gorm.DB, projectName string, errInternal, errNotFound error) *QB {
	return &QB{
		db:          db,
		projectName: projectName,
		errInternal: errInternal,
		errNotFound: errNotFound,
	}
}
