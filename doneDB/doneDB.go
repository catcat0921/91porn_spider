package doneDB

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"spider91/catch"
)

type VideoDB struct {
	db *sql.DB
}

func (v *VideoDB) Close() error {
	return v.db.Close()
}

func (v *VideoDB) AddDone(vis []*catch.VideoInfo) (fails []*catch.VideoInfo) {

	for _, vi := range vis {
		stmt, err := v.db.Prepare("INSERT INTO done(viewkey, title, owner, UpTime)  values(?, ?, ?, ?)")
		if err == nil {
			_, err := stmt.Exec(vi.ViewKey, vi.Title, vi.Owner, vi.UpTime)
			//fmt.Println(err)
			if err == nil {
				continue
			}
		}

		fails = append(fails, vi)
	}

	return
}

func (v *VideoDB) DelRepeat(vis []*catch.VideoInfo) (pick []*catch.VideoInfo) {
	for _, vi := range vis {
		var count int64
		err := v.db.QueryRow("select count(*)FROM done WHERE viewkey=?", vi.ViewKey).Scan(&count)
		if count == 0 && err == nil {
			pick = append(pick, vi)
		}
	}

	return
}

func (v *VideoDB) UpdateUD(vis []*catch.VideoInfo) (fails []*catch.VideoInfo) {

	for _, vi := range vis {
		var count int64
		err := v.db.QueryRow("select count(*)FROM undone WHERE viewkey=?", vi.ViewKey).Scan(&count)
		if count == 0 && err == nil {
			stmt, err := v.db.Prepare("INSERT INTO undone(viewkey, title, owner, UpTime, failcount)  values(?, ?, ?, ?, ?)")
			if err == nil {
				_, err := stmt.Exec(vi.ViewKey, vi.Title, vi.Owner, vi.UpTime, 1)
				//fmt.Println(err)
				if err == nil {
					continue
				}
			}
		} else if count > 0 && err == nil {
			failcount := 0
			err := v.db.QueryRow("select failcount FROM undone WHERE viewkey=?", vi.ViewKey).Scan(&failcount)
			if err == nil {
				if failcount >= 3 {
					//del item
					stmt, err := v.db.Prepare("delete from undone where viewkey=?")
					if err == nil {
						stmt.Exec(vi.ViewKey)
						continue
					}
				} else {
					//update item
					stmt, err := v.db.Prepare("update undone set failcount=? where viewkey=?")
					if err == nil {
						stmt.Exec(failcount+1, vi.ViewKey)
						continue
					}
				}
			}
		}

		fails = append(fails, vi)
	}

	return
}

func (v *VideoDB) GetUD() (undone []*catch.VideoInfo) {

	rows, err := v.db.Query("SELECT viewkey, title, owner, UpTime FROM undone")
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		vi := catch.VideoInfo{}
		err := rows.Scan(&vi.ViewKey, &vi.Title, &vi.Owner, &vi.UpTime)
		if err == nil {
			undone = append(undone, &vi)
		}
	}

	return
}

func OpenVDB(filename string) (vdb *VideoDB, err error) {
	vdb = new(VideoDB)
	vdb.db, err = sql.Open("sqlite3", filename)
	if err == nil {

		sql_table := `
		CREATE TABLE IF NOT EXISTS "done" (
			"viewkey" VARCHAR(64) PRIMARY KEY,
			"title" VARCHAR(64) NULL,
		    "owner" VARCHAR(64) NULL,
			"UpTime" TIMESTAMP default (datetime('now', 'localtime'))  
		);
		CREATE TABLE IF NOT EXISTS "undone" (
			"viewkey" VARCHAR(64) PRIMARY KEY,
			"title" VARCHAR(64) NULL,
		    "owner" VARCHAR(64) NULL,
			"UpTime" TIMESTAMP default (datetime('now', 'localtime')),
			"failcount" INTEGER NOT NULL DEFAULT 0
		);
	   `
		var res sql.Result
		res, err = vdb.db.Exec(sql_table)
		fmt.Println(res, err)
		return
	}

	return nil, err
}
