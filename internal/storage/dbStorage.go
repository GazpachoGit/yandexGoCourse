package storage

import (
	"database/sql"
	"errors"
	"log"

	"github.com/GazpachoGit/yandexGoCourse/internal/model"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const (
	ErrNotFound = "can't find id"
)

type PgDb struct {
	dbConn            *sqlx.DB
	sqlSelectURL      *sqlx.Stmt
	sqlInsertURL      *sqlx.Stmt
	sqlSelectUserURLs *sqlx.Stmt
}

func InitDb(psqlInfo string) (*PgDb, error) {
	p := &PgDb{nil, nil, nil, nil}

	db, err := sqlx.Connect("postgres", psqlInfo)
	if err != nil {
		return p, err
	}
	p.dbConn = db

	if err = p.dbConn.Ping(); err != nil {
		return p, err
	}

	//create table
	if err = p.createTables(); err != nil {
		return p, err
	}

	//create statements
	if err = p.createSqlStmts(); err != nil {
		return p, err
	}

	return p, nil

}

func (p *PgDb) PingDB() error {
	if err := p.dbConn.Ping(); err != nil {
		return err
	}
	return nil
}

func (p *PgDb) Close() {
	if p.dbConn != nil {
		if err := p.dbConn.Close(); err != nil {
			log.Fatalln(err)
		}
	}
	if p.sqlInsertURL != nil {
		if err := p.sqlInsertURL.Close(); err != nil {
			log.Fatalln(err)
		}
	}
	if p.sqlSelectURL != nil {
		if err := p.sqlSelectURL.Close(); err != nil {
			log.Fatalln(err)
		}
	}
}

func (p *PgDb) createTables() error {
	create_sql := ` CREATE TABLE IF NOT EXISTS public.urls_torn (
		id SERIAL NOT NULL PRIMARY KEY,
       	original_url TEXT NOT NULL,
	   	user_id TEXT NOT NULL);
    `
	if _, err := p.dbConn.Exec(create_sql); err != nil {
		return err
	}
	return nil
}

func (p *PgDb) createSqlStmts() error {
	if stmt, err := p.dbConn.Preparex("INSERT INTO public.urls_torn(original_url, user_id) VALUES($1, $2) RETURNING id"); err != nil {
		return err
	} else {
		p.sqlInsertURL = stmt
	}

	if stmt, err := p.dbConn.Preparex("SELECT original_url FROM public.urls_torn WHERE id = $1 LIMIT 1"); err != nil {
		return err
	} else {
		p.sqlSelectURL = stmt
	}

	if stmt, err := p.dbConn.Preparex("SELECT id, original_url FROM public.urls_torn WHERE user_id = $1"); err != nil {
		return err
	} else {
		p.sqlSelectUserURLs = stmt
	}
	return nil
}

func (p *PgDb) Set(original_url string, user string) (int, error) {
	var id int
	if err := p.sqlInsertURL.QueryRowx(original_url, user).Scan(&id); err != nil {
		return 0, err
	} else {
		return int(id), nil
	}

}

func (p *PgDb) Get(id int) (string, error) {
	var original_url string
	row := p.sqlSelectURL.QueryRowx(id)
	err := row.Scan(&original_url)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New(ErrNotFound)
		}
		return "", err
	}
	return original_url, nil
}

func (p *PgDb) GetUserURLs(user string) ([]model.StorageURLInfo, error) {
	var URLs []model.StorageURLInfo
	if err := p.sqlSelectUserURLs.Select(&URLs, user); err != nil {
		return nil, err
	}
	return URLs, nil
}
func (p *PgDb) SetBatchURLs(input *[]*model.HandlerURLInfo, username string) (*map[string]*model.StorageURLInfo, error) {
	tx, err := p.dbConn.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	output := make(map[string]*model.StorageURLInfo)
	for _, v := range *input {
		if v.Correlation_id == "" {
			return nil, errors.New("empty correlation")
		}
		if _, ok := output[v.Correlation_id]; ok {
			return nil, errors.New("dublicate correlation")
		}
		id, err := p.Set(v.Original_url, username)
		if err != nil {
			return nil, err
		}
		output[v.Correlation_id] = &model.StorageURLInfo{
			Id:           id,
			Original_url: v.Original_url,
		}
	}
	return &output, tx.Commit()
}

// func main() {
// 	db, err := InitDb()
// 	defer db.Close()
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
// 	if _, err = db.Set("123", "321"); err != nil {
// 		fmt.Println(err)
// 	}
// 	if res, err := db.Get(1); err != nil {
// 		fmt.Println(err)

// 	} else {
// 		fmt.Println(res)
// 	}

// }
