# queryBuilder
Abstraction over the gorm.io 

readme wip until first version isn't tested

Expected Usage:

1) init model with query builder

```go
func New(connURL string) *Model {
	postgres.New(postgres.Config{})
	db, err := gorm.Open(postgres.Open(connURL), &gorm.Config{})
	if err != nil {
		logrus.WithError(err).Fatal("can't connect to database")
	}
	return &Model{
		db: db,
		QB: queryBuilder.New(db, "business_network", common.ErrInternal, common.ErrNotFound),
	}
}
```

2) Use qb on top-tier for make common simply queries without worries about tracing and logging errors

```go
func (a *Application) DoSomethingWithUsers(filter model.User) error {
    var users []model.User
    if err := a.model.Where(filter).Find(&users); err != nil {
        return err
    }
    // do something
}
```

// FIXME: Does not work as expected.
// TODO: Find solution for use custom project-related methods. 
```go
func (a *Application) DoSomethingWithUsers(filter model.User, limit int) error {
    users, err := a.model.Preload("Friends").CustomFindMethod(&users)
	if err != nil {
        return err
    }
    // do something
}
```