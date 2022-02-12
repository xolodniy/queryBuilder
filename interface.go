package queryBuilder

// QueryBuilder expands default gorm methods
// there are embedded logging, common errors and little more simply signature
type QueryBuilder interface {
	Preload(column string, conditions ...interface{}) *QB
	Debug() *QB
	Unscoped() *QB
	IgnoreConflicts() *QB
	Model(value interface{}) *QB
	Select(query interface{}, args ...interface{}) *QB
	Table(name string) *QB
	Limit(limit int) *QB
	Offset(offset int) *QB
	Order(value interface{}) *QB
	Set(name string, value interface{}) *QB
	Pluck(column string, value interface{}) error
	First(out interface{}, where ...interface{}) error
	Last(out interface{}, where ...interface{}) error
	Find(out interface{}, where ...interface{}) error
	Scan(dest interface{}) error
	Create(value interface{}) error
	Save(value interface{}) error
	Omit(value ...string) *QB
	Updates(attrs interface{}) error
	Delete(value interface{}, where ...interface{}) error
	Where(query interface{}, args ...interface{}) *QB
	Count() (int64, error)
	Not(query interface{}, args ...interface{}) *QB
	Group(name string) *QB
	Having(query interface{}, args ...interface{}) *QB
	Take(dest interface{}, conds ...interface{}) error
	BatchFind(dest interface{}, batchSize int, fc func(tx *QB, batch int) error) error
	Joins(query string, args ...interface{}) *QB
	UpdateByFilter(filter interface{}, values interface{}) error

	exec(sql string, values ...interface{}) error
}
