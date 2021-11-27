package upload

import (
	"bytes"

	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx"
	"github.com/scylladb/gocqlx/qb"
	"github.com/scylladb/gocqlx/table"
	"go.uber.org/zap"
)

type scyllaRepo struct {
	log  *zap.Logger
	conn *gocql.Session
	stmt *statements
}

func NewScyllaRepository(logger *zap.Logger, session *gocql.Session) Repository {
	return &scyllaRepo{
		log:  logger,
		conn: session,
		stmt: createStatements(),
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

type File struct {
	ID              string `db:"id"`
	FileName        string `db:"file_name"`
	FileType        string `db:"type"`
	TotalSize       int    `db:"size"`
	File_data       []byte `db:"file_data"`
	FIle_DataBuffer bytes.Buffer
}

func (r *scyllaRepo) InsertFile(f File) error {
	err := gocqlx.Query(r.conn.Query(r.stmt.ins.stmt), r.stmt.ins.names).BindStruct(f).ExecRelease()
	return err
}

func createStatements() *statements {
	m := table.Metadata{
		Name:    "upload_file",
		Columns: []string{"id", "file_name", "type", "size", "file_data"},
		PartKey: []string{"id", "type"},
	}
	tbl := table.New(m)

	deleteStmt, deleteNames := tbl.Delete()
	insertStmt, insertNames := tbl.Insert()
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
