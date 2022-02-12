package queryBuilder

import (
	"errors"
	"fmt"
	"gorm.io/gorm/clause"
	"reflect"
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/xolodniy/pretty"
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

// Preload is gorm interface func
func (qb *QB) Preload(column string, conditions ...interface{}) *QB {
	trace := initLogTrace(qb.logTrace)
	trace["preloadColumn-"+column] = column
	trace["preloadConditions-"+column] = conditions
	return &QB{db: qb.db.Preload(column, conditions...), logTrace: trace}
}

// Debug is gorm interface func
func (qb *QB) Debug() *QB {
	return &QB{db: qb.db.Debug(), logTrace: qb.logTrace}
}

// Unscoped is gorm interface func
func (qb *QB) Unscoped() *QB {
	trace := initLogTrace(qb.logTrace)
	trace["unscoped"] = true
	return &QB{db: qb.db.Unscoped(), logTrace: qb.logTrace}
}

// Model is gorm interface func
func (qb *QB) Model(value interface{}) *QB {
	trace := initLogTrace(qb.logTrace)
	trace["QBValueType"] = pretty.Print(value)
	return &QB{db: qb.db.Model(value), logTrace: trace}
}

// Select is gorm interface func
func (qb *QB) Select(query interface{}, args ...interface{}) *QB {
	trace := initLogTrace(qb.logTrace)
	trace["selectQuery"] = query
	trace["selectArgs"] = pretty.Print(args)
	return &QB{db: qb.db.Select(query, args...), logTrace: trace}
}

// Table is gorm interface func
func (qb *QB) Table(name string) *QB {
	trace := initLogTrace(qb.logTrace)
	trace["tableName"] = name
	return &QB{db: qb.db.Table(name), logTrace: trace}
}

// Limit is gorm interface func
func (qb *QB) Limit(limit int) *QB {
	trace := initLogTrace(qb.logTrace)
	trace["limit"] = limit
	return &QB{db: qb.db.Limit(limit), logTrace: trace}
}

// Offset is gorm interface func
func (qb *QB) Offset(offset int) *QB {
	trace := initLogTrace(qb.logTrace)
	trace["offset"] = offset
	return &QB{db: qb.db.Offset(offset), logTrace: trace}
}

// Order is gorm interface func
func (qb *QB) Order(value interface{}) *QB {
	trace := initLogTrace(qb.logTrace)
	trace["order"] = pretty.Print(value)
	return &QB{db: qb.db.Order(value), logTrace: trace}
}

// Joins is gorm interface func
func (qb *QB) Joins(query string, args ...interface{}) *QB {
	trace := initLogTrace(qb.logTrace)
	var i int
	for {
		if _, ok := trace["joinsQuery"+strconv.Itoa(i)]; !ok {
			break
		}
		i++
	}
	trace["joinsQuery"+strconv.Itoa(i)] = query
	if len(args) > 0 {
		trace["joinsArgs"+strconv.Itoa(i)] = pretty.Print(args)
	}
	return &QB{db: qb.db.Joins(query, args), logTrace: trace}
}

func (qb *QB) Set(name string, value interface{}) *QB {
	trace := initLogTrace(qb.logTrace)
	var i int
	for {
		if _, ok := trace["setName"+strconv.Itoa(i)]; !ok {
			break
		}
		i++
	}
	trace["setName"+strconv.Itoa(i)] = name
	trace["setValue"+strconv.Itoa(i)] = value
	return &QB{db: qb.db.Set(name, value), logTrace: trace}
}
func (qb *QB) IgnoreConflicts() *QB {
	trace := initLogTrace(qb.logTrace)
	trace["ignoreConflicts"] = true
	return &QB{db: qb.db.Clauses(clause.OnConflict{DoNothing: true}), logTrace: trace}
}

// Pluck is gorm interface func
func (qb *QB) Pluck(column string, value interface{}) error {
	err := qb.db.Pluck(column, value).Error
	if err != nil {
		logrus.WithError(err).WithFields(qb.logTrace).WithFields(logrus.Fields{
			"typeOfPluckingValue": fmt.Sprintf("%T", value),
			"pluckColumnName":     column,
			"trace":               qb.GetFrames(),
		}).Error("can't pluck object from the database")
		return qb.errInternal
	}
	return nil
}

// First is gorm interface func
func (qb *QB) First(out interface{}, where ...interface{}) error {
	err := qb.db.First(out, where...).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return qb.errNotFound
	}
	if err != nil {
		logFields := logrus.Fields{
			"trace":    qb.GetFrames(),
			"firstOut": pretty.Print(out),
		}
		if len(where) > 0 {
			logFields["firstWhere"] = pretty.Print(where)
		}
		logrus.WithError(err).WithFields(qb.logTrace).WithFields(logFields).Error("can't get first object from the database")
		return qb.errInternal
	}
	return nil
}

// Last is gorm interface func
func (qb *QB) Last(out interface{}, where ...interface{}) error {
	err := qb.db.Last(out, where...).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return qb.errNotFound
	}
	if err != nil {
		logFields := logrus.Fields{
			"trace":   qb.GetFrames(),
			"lastOut": pretty.Print(out),
		}
		if len(where) > 0 {
			logFields["lastWhere"] = pretty.Print(where)
		}
		logrus.WithError(err).WithFields(qb.logTrace).WithFields(logFields).Error("can't get last object from the database")
		return qb.errInternal
	}
	return nil
}

// Take is gorm interface func
func (qb *QB) Take(dest interface{}, conditions ...interface{}) error {
	err := qb.db.Take(dest, conditions...).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return qb.errNotFound
	}
	if err != nil {
		logFields := logrus.Fields{
			//"takeOutValue":       fmt.Sprintf("%+v", dest),
			//"takeTypeOfOutValue": fmt.Sprintf("%T", dest),
			"takeWhereCondition": fmt.Sprintf("%+v", conditions),
			"takeDest":           pretty.Print(dest),
			"trace":              qb.GetFrames(),
		}
		if len(conditions) > 0 {
			logFields["takeConditions"] = pretty.Print(conditions)
		}
		logrus.WithError(err).WithFields(qb.logTrace).WithFields(logFields).Error("can't take object from the database")
		return qb.errInternal
	}
	return nil
}

// Find is gorm interface func
func (qb *QB) Find(out interface{}, where ...interface{}) error {
	err := qb.db.Find(out, where...).Error
	if err != nil {
		logFields := logrus.Fields{
			//"findOutValue":       fmt.Sprintf("%+v", out),
			//"findTypeOfOutValue": fmt.Sprintf("%T", out),
			"findOut": pretty.Print(out),
			//"findWhereCondition": fmt.Sprintf("%+v", where),
			"trace": qb.GetFrames(),
		}
		if len(where) > 0 {
			logFields["findWhere"] = pretty.Print(where)
		}
		logrus.WithError(err).WithFields(qb.logTrace).WithFields(logFields).Error("can't find from the database")
		return qb.errInternal
	}
	return nil
}

// Scan is gorm interface func
func (qb *QB) Scan(dest interface{}) error {
	err := qb.db.Scan(dest).Error
	if err != nil {
		logrus.WithError(err).WithFields(qb.logTrace).WithFields(logrus.Fields{
			//"scanOutValue":       fmt.Sprintf("%+v", dest),
			//"scanTypeOfOutValue": fmt.Sprintf("%T", dest),
			"scanDest": pretty.Print(dest),
			"trace":    qb.GetFrames(),
		}).Error("can't scan from the database")
		return qb.errInternal
	}
	return nil
}

// Create is gorm interface func
func (qb *QB) Create(value interface{}) error {
	err := qb.db.Create(value).Error
	if err != nil {
		logrus.WithError(err).WithFields(qb.logTrace).WithFields(logrus.Fields{
			//"createValue":       fmt.Sprintf("%+v", value),
			//"createTypeOfValue": fmt.Sprintf("%T", value),
			"createValue": pretty.Print(value),
			"trace":       qb.GetFrames(),
		}).Error("can't create value in database")
		return qb.errInternal
	}
	return nil
}

// Save is gorm interface func
func (qb *QB) Save(value interface{}) error {
	if err := qb.db.Save(value).Error; err != nil {
		logrus.WithError(err).WithFields(qb.logTrace).WithFields(logrus.Fields{
			//"savedValue":     fmt.Sprintf("%+v", value),
			//"savedValueType": fmt.Sprintf("%T", value),
			"saveValue": pretty.Print(value),
			"trace":     qb.GetFrames(),
		}).Error("can't save object in a database")
		return qb.errInternal
	}
	return nil
}

// Omit is gorm interface func
func (qb *QB) Omit(value ...string) *QB {
	trace := initLogTrace(qb.logTrace)
	trace["omit"] = value
	return &QB{db: qb.db.Omit(value...), logTrace: trace}
}

// Updates is gorm interface func
func (qb *QB) Updates(attrs interface{}) error {
	if err := qb.db.Updates(attrs).Error; err != nil {
		logrus.WithError(err).WithFields(qb.logTrace).WithFields(logrus.Fields{
			//"updateAttrs":     fmt.Sprintf("%+v", attrs),
			//"updateAttrsType": fmt.Sprintf("%T", attrs),
			"updateAttrs": pretty.Print(attrs),
			"trace":       qb.GetFrames(),
		}).Error("can't update object in database")
		return qb.errInternal
	}
	return nil
}

// Delete is gorm interface func
func (qb *QB) Delete(value interface{}, where ...interface{}) error {
	if err := qb.db.Delete(value, where...).Error; err != nil {
		logFields := logrus.Fields{
			//"deleteValue": fmt.Sprintf("%+v", value),
			//"deleteWhere": fmt.Sprintf("%+v", where),
			"deleteValue": pretty.Print(value),
			"trace":       qb.GetFrames(),
		}
		if len(where) > 0 {
			logFields["deleteWhere"] = pretty.Print(where)
		}
		logrus.WithError(err).WithFields(qb.logTrace).WithFields(logFields).Error("can't delete object from DB")
		return qb.errInternal
	}
	return nil
}

// Where is gorm interface func
func (qb *QB) Where(query interface{}, args ...interface{}) *QB {
	trace := initLogTrace(qb.logTrace)
	var i int
	for {
		if _, ok := trace["whereQuery"+strconv.Itoa(i)]; !ok {
			break
		}
		i++
	}
	trace["whereQuery"+strconv.Itoa(i)] = pretty.Print(query)
	//trace["whereQueryType"] = fmt.Sprintf("%T", query)
	if len(args) > 0 {
		trace["whereArgs"+strconv.Itoa(i)] = pretty.Print(args)
	}
	return &QB{db: qb.db.Where(query, args...), logTrace: trace}
}

// Count is gorm interface func
func (qb *QB) Count() (int64, error) {
	var c int64
	if err := qb.db.Count(&c).Error; err != nil {
		logrus.WithError(err).WithFields(qb.logTrace).WithFields(logrus.Fields{
			"trace": qb.GetFrames(),
		}).Error("can't count objects in DB")
		return 0, qb.errInternal
	}
	return c, nil
}

// Not is gorm interface func
func (qb *QB) Not(query interface{}, args ...interface{}) *QB {
	trace := initLogTrace(qb.logTrace)
	trace["notQuery"] = pretty.Print(query)
	if len(args) > 0 {
		trace["notArgs"] = args
	}
	return &QB{db: qb.db.Not(query, args...), logTrace: trace}
}

// Group is gorm interface func
func (qb *QB) Group(name string) *QB {
	trace := initLogTrace(qb.logTrace)
	trace["groupName"] = name
	return &QB{db: qb.db.Group(name), logTrace: trace}
}

// Having is gorm interface func
func (qb *QB) Having(query interface{}, args ...interface{}) *QB {
	trace := initLogTrace(qb.logTrace)
	trace["havingQuery"] = query
	trace["havingArgs"] = args
	return &QB{db: qb.db.Having(query, args...), logTrace: trace}
}

func (qb *QB) exec(sql string, values ...interface{}) error {
	if err := qb.db.Exec(sql, values...).Error; err != nil {
		logrus.WithError(err).WithFields(qb.logTrace).WithFields(logrus.Fields{
			"trace":      qb.GetFrames(),
			"execSql":    sql,
			"execValues": values,
		}).Error("can't exec sql in DB")
		return qb.errInternal
	}
	return nil
}

func (qb *QB) raw(sql string, values ...interface{}) *QB {
	trace := initLogTrace(qb.logTrace)
	trace["rawSql"] = sql
	trace["rawValues"] = values
	return &QB{db: qb.db.Raw(sql, values...), logTrace: trace}
}

// BatchFind is gorm interface func
// FIXME: does not works.
// got error "primary key required" when tried to fetch user followers
// maybe it composite key relates?
func (qb *QB) BatchFind(dest interface{}, batchSize int, fc func(tx *QB, batch int) error) error {
	err := qb.db.FindInBatches(dest, batchSize, func(tx *gorm.DB, batch int) error {
		return fc(qb, batch)
	}).Error
	if err != nil {
		logFields := logrus.Fields{
			"batchFindDest": pretty.Print(dest),
			"batchSize":     batchSize,
			"trace":         qb.GetFrames(),
		}
		logrus.WithError(err).WithFields(qb.logTrace).WithFields(logFields).Error("can't find from the database")
		return qb.errInternal
	}
	return nil
}

// UpdateByFilter is gorm extension. Allow omitting .Model() and .Where() methods
func (qb *QB) UpdateByFilter(filter interface{}, values interface{}) error {
	if reflect.DeepEqual(filter, reflect.Zero(reflect.TypeOf(filter)).Interface()) {
		logrus.Error("queryBuilder.UpdateByFilter called for empty filter")
		return qb.errInternal
	}
	if err := qb.db.Model(filter).Where(filter).Updates(values).Error; err != nil {
		logrus.WithError(err).WithFields(qb.logTrace).WithFields(logrus.Fields{
			"UpdateByFilterFilter":     fmt.Sprintf("%+v", filter),
			"UpdateByFilterFilterType": fmt.Sprintf("%T", filter),
			"UpdateByFilterValues":     fmt.Sprintf("%+v", values),
			"UpdateByFilterValuesType": fmt.Sprintf("%T", values),
			"trace":                    qb.GetFrames(),
		}).Error("can't update object in database")
		return qb.errInternal
	}
	return nil
}
