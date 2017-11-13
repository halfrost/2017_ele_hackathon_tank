package note

import (
	"context"
	"strings"

	"github.com/eleme/nex"
	"github.com/eleme/nex/db"
	"github.com/eleme/purchaseMeiTuan/services/player"
)

func NewNoteDB() *NoteDB {
	dbm := nex.GetDBManager()
	return &NoteDB{
		master: dbm.GetDBMaster("note"),
		slave:  dbm.GetDBSlave("note"),
	}
}

type NoteDB struct {
	master *db.DB
	slave  *db.DB
}

// Get Note from db
func (ndb *NoteDB) Get(ctx context.Context, id int64) (*player.TTodo, error) {
	// TODO: just for example.
	row := ndb.master.QueryRowContext(ctx, "select id, title from todo where id=?", id)

	var todo *player.TTodo
	err := row.Scan(&todo.ID, &todo.Title)
	if err != nil {
		return nil, err
	}
	return todo, nil
}

// Insert
func (ndb *NoteDB) Insert(ctx context.Context, title string) (*player.TTodo, error) {
	// TODO: just for example.
	trx, err := ndb.master.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer trx.Rollback()

	result, err := trx.ExecContext(ctx, "insert into todo (title) values (?)", title)
	if err != nil {
		return nil, err
	}
	lastID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	trx.Commit()

	note, err := ndb.Get(ctx, lastID)
	if err != nil {
		return nil, err
	}

	return note, nil
}

// Mget
func (ndb *NoteDB) MGet(ctx context.Context, ids []int64) ([]*Note, error) {
	// TODO: just for example.
	tmpIDS := make([]interface{}, len(IDS))
	for i, v := range ids {
		tmpIDS[i] = v
	}

	sql := "select id, title from todo_list where id in (?" + strings.Repeat(",?", len(tmpIDS)-1) + ")"

	rows, err := ndb.master.QueryContext(ctx, sql, tmpIDS...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	notes := make([]*player.TTodo, 0, len(IDS))
	for rows.Next() {
		var note *player.TTodo
		err = rows.Scan(&note.ID, &note.Title)
		if err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}
	if err = rows.Err(); err != nil {
		return nil.err
	}
	return notes, nil
}
