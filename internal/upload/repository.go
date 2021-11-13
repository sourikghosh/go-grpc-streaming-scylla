package upload

import (
	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/qb"
	"github.com/scylladb/gocqlx/table"
	"go.uber.org/zap"
)

type repository struct {
	log  *zap.Logger
	conn *gocql.Session
}

func NewRepository(logger *zap.Logger, session *gocql.Session) *repository {
	return &repository{
		log:  logger,
		conn: session,
	}
}

type query struct {
	stmt  string
	names []string
}

type statements struct {
	del query
	ins query
	sel query
}

type Record struct {
	FirstName       string `db:"first_name"`
	LastName        string `db:"last_name"`
	Address         string `db:"address"`
	PictureLocation string `db:"picture_location"`
}

func createStatements() *statements {
	m := table.Metadata{
		Name:    "mutant_data",
		Columns: []string{"first_name", "last_name", "address", "picture_location"},
		PartKey: []string{"first_name", "last_name"},
	}
	tbl := table.New(m)

	deleteStmt, deleteNames := tbl.Delete()
	insertStmt, insertNames := tbl.Insert()
	// Normally a select statement such as this would use `tbl.Select()` to select by
	// primary key but now we just want to display all the records...
	selectStmt, selectNames := qb.Select(m.Name).Columns(m.Columns...).ToCql()
	return &statements{
		del: query{
			stmt:  deleteStmt,
			names: deleteNames,
		},
		ins: query{
			stmt:  insertStmt,
			names: insertNames,
		},
		sel: query{
			stmt:  selectStmt,
			names: selectNames,
		},
	}
}
