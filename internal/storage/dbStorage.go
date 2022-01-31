package storage

import (
	"database/sql"
	"errors"
	"log"

	myerrors "github.com/GazpachoGit/yandexGoCourse/internal/errors"
	"github.com/GazpachoGit/yandexGoCourse/internal/model"
	"github.com/jackc/pgerrcode"
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
       	original_url TEXT NOT NULL UNIQUE,
	   	user_id TEXT NOT NULL);
    `
	// if _, err := p.dbConn.Exec("DROP TABLE IF EXISTS public.urls_torn"); err != nil {
	// 	return err
	// }
	if _, err := p.dbConn.Exec(create_sql); err != nil {
		return err
	}
	return nil
}

func (p *PgDb) createSqlStmts() error {
	insertSQL := `with stmt AS (INSERT INTO public.urls_torn(original_url, user_id)
	VALUES ($1, $2) 
	ON CONFLICT(original_url) do nothing
	RETURNING id, false as conf)

	select id, conf from stmt 
	where id is not null
	UNION ALL
	select id, true from public.urls_torn
	where original_url = $1 and not exists (select 1 from stmt)`

	if stmt, err := p.dbConn.Preparex(insertSQL); err != nil {
		return err
	} else {
		p.sqlInsertURL = stmt
	}

	if stmt, err := p.dbConn.Preparex("SELECT original_url FROM public.urls_torn WHERE id = $1 LIMIT 1"); err != nil {
		return err
	} else {
		p.sqlSelectURL = stmt
	}

	if stmt, err := p.dbConn.Preparex("SELECT id, original_url,user_id FROM public.urls_torn WHERE user_id = $1"); err != nil {
		return err
	} else {
		p.sqlSelectUserURLs = stmt
	}
	return nil
}

func (p *PgDb) SetURL(original_url string, user string) (int, error) {
	if insertInfo, err := p.Set(original_url, user); err != nil {
		return 0, err
	} else {
		if insertInfo.Conf {
			return insertInfo.ID, myerrors.NewInsertConflictError([]string{original_url}, errors.New(pgerrcode.UniqueViolation))
		}
		return insertInfo.ID, nil
	}

}

func (p *PgDb) Set(original_url string, user string) (*model.StorageInsertInfo, error) {
	var insertInfo model.StorageInsertInfo
	if err := p.sqlInsertURL.QueryRowx(original_url, user).StructScan(&insertInfo); err != nil {
		return nil, err
	} else {
		return &insertInfo, nil
	}
}

func (p *PgDb) GetURL(id int) (string, error) {
	var original_url string
	row := p.sqlSelectURL.QueryRowx(id)
	err := row.Scan(&original_url)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", myerrors.NewNotFoundError()
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
	if URLs == nil {
		return nil, myerrors.NewNotFoundError()
	}
	log.Println("userId for the first: ", URLs[0].UserID)
	return URLs, nil
}
func (p *PgDb) SetBatchURLs(input *[]*model.HandlerURLInfo, username string) (*map[string]*model.StorageURLInfo, error) {
	tx, err := p.dbConn.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	output := make(map[string]*model.StorageURLInfo)
	confOriginURL := make([]string, 0)
	for _, v := range *input {
		if v.CorrelationID == "" {
			return nil, errors.New("empty correlation")
		}
		insertInfo, err := p.Set(v.OriginalURL, username)
		if err != nil {
			return nil, err
		}
		if insertInfo.Conf {
			confOriginURL = append(confOriginURL, v.OriginalURL)
		}
		output[v.CorrelationID] = &model.StorageURLInfo{
			ID:          insertInfo.ID,
			OriginalURL: v.OriginalURL,
		}
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	if len(confOriginURL) != 0 {
		err = myerrors.NewInsertConflictError(confOriginURL, errors.New(pgerrcode.UniqueViolation))
		return &output, err
	}
	return &output, nil
}
