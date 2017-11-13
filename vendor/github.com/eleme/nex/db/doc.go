// Package db is responsible for logics interacting with mysql.
//
// Database configs
//
// DB reads configs from huskar with config key `DBSettings`, its format is:
//  {
//  	"name": {
//          "driver": "mysql",            # Only 'mysql' and 'postgres' supported, default is 'mysql',
//  		"master": "uri of master db", # Note nex depends on DAL for read/write sepration
//  		"slave": "uri of slave db",   # Note nex depends on DAL for read/write sepration
//  		"max_lifetime": 300,          # The max lifetime of a connection, default: 300
//  		"max_open_conns": 500,        # The max opened connections maintained in connection pool, default: 500
//  		"max_idle_conns": 300,        # The max idle connections maintained in connection pool, default: 300
//  		"enable_log": false,          # Whether log every mysql execution, default: false
//  		"enable_metrics": false       # Whether send metrics with every mysql execution, default: false
//  	},
//  }
//
// Uri should follow this format
//   user:password@tcp(ip:port)/db
// Such as:
//   root:my-secret-pw@tcp(127.0.0.1:3306)/note
//
// Examples
//
// This example shows how to get a single row from db.
//   func QuerySingleRow(ctx context.Context, ID int64) (int64, string, error) {
//   	row := nex.GetDBManager().GetDBMaster("note").QueryRowContext(ctx, "select id, title from todo_list where id=?", ID)
//   	var (
//   		id   int64
//   		name string
//   	)
//   	if err := row.Scan(&id, &name); err != nil {
//   		return 0, "", err
//   	}
//   	return id, name, nil
//   }
//
// This example shows how to query many rows from databases.
//   func QueryIdsGreaterThan(ctx context.Context, ID int) ([]int, error) {
//   	sql := "select id from todo_list where id > ?"
//   	rows, err := nex.GetDBManager().GetDBMaster("note").QueryContext(ctx, sql, ID)
//   	if err != nil {
//   		return nil, err
//   	}
//      defer rows.Close()
//   	var (
//   		id  int
//   		IDs []int
//   	)
//   	for rows.Next() {
//   		if err = rows.Scan(&id); err != nil {
//   			return nil, err
//   		}
//   		IDs = append(IDs, id)
//   	}
//      if err = rows.Err(); err != nil {
//          return nil, err
//      }
//   	return IDs, nil
//   }
//
// This example shows how to insert a record into database.
// You must use transaction to do all data modifying operations,
// or risk data data corruption directly using `DB.execContext`
//   func Insert(ctx context.Context, Name string) (int64, error) {
//   	trx, err := nex.GetDBManager().GetDBMaster("note").BeginTx(ctx, nil)
//   	if err != nil {
//   		return 0, err
//   	}
//   	defer trx.Rollback()
//   	result, err := trx.ExecContext(ctx, "insert into todo_list (title) values (?)", Name)
//   	if err != nil {
//   		return 0, err
//   	}
//   	lastID, err := result.LastInsertId()
//   	if err != nil {
//   		return 0, err
//   	}
//  	if err = trx.Commit(); err != nil {
//  		return 0, err
//  	}
//   	return lastID, nil
//   }
package db
