package storage

import (
	"database/sql"
	"errors"

	myerrors "github.com/GazpachoGit/yandexGoCourse/internal/errors"
	"github.com/GazpachoGit/yandexGoCourse/internal/model"
	"github.com/jackc/pgerrcode"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const (
	ErrNotFound = "can't find id"
)

type PgDB struct {
	dbConn            *sqlx.DB
	sqlSelectURL      *sqlx.Stmt
	sqlInsertURL      *sqlx.Stmt
	sqlSelectUserURLs *sqlx.Stmt
}

func InitDB(psqlInfo string) (*PgDB, error) {
	p := &PgDB{nil, nil, nil, nil}

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
	if err = p.createSQLStmts(); err != nil {
		return p, err
	}

	return p, nil

}

func ConfigDBForTest(db *sqlx.DB) (*PgDB, error) {
	p := &PgDB{nil, nil, nil, nil}
	p.dbConn = db

	if err := p.dbConn.Ping(); err != nil {
		return p, err
	}

	//create statements
	if err := p.createSQLStmts(); err != nil {
		return p, err
	}

	return p, nil
}

func (p *PgDB) PingDB() error {
	if err := p.dbConn.Ping(); err != nil {
		return err
	}
	return nil
}

func (p *PgDB) Close() error {
	if p.dbConn != nil {
		if err := p.dbConn.Close(); err != nil {
			return err
		}
	}
	if p.sqlInsertURL != nil {
		if err := p.sqlInsertURL.Close(); err != nil {
			return err
		}
	}
	if p.sqlSelectURL != nil {
		if err := p.sqlSelectURL.Close(); err != nil {
			return err
		}
	}
	if p.sqlSelectUserURLs != nil {
		if err := p.sqlSelectUserURLs.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (p *PgDB) createTables() error {
	createSQL := ` CREATE TABLE IF NOT EXISTS public.urls_torn (
		id SERIAL NOT NULL PRIMARY KEY,
       	original_url TEXT NOT NULL UNIQUE,
	   	user_id TEXT NOT NULL);
    `
	if _, err := p.dbConn.Exec(createSQL); err != nil {
		return err
	}
	return nil
}

func (p *PgDB) createSQLStmts() error {
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

	selectOneSQL := "SELECT original_url FROM public.urls_torn WHERE id = $1 LIMIT 1"
	if stmt, err := p.dbConn.Preparex(selectOneSQL); err != nil {
		return err
	} else {
		p.sqlSelectURL = stmt
	}
	selectUserURLsSQL := "SELECT id, original_url FROM public.urls_torn WHERE user_id = $1"
	if stmt, err := p.dbConn.Preparex(selectUserURLsSQL); err != nil {
		return err
	} else {
		p.sqlSelectUserURLs = stmt
	}
	return nil
}

func (p *PgDB) SetURL(originalURL string, user string) (int, error) {
	if insertInfo, err := p.Set(originalURL, user); err != nil {
		return 0, err
	} else {
		if insertInfo.Conf {
			return insertInfo.ID, myerrors.NewInsertConflictError([]string{originalURL}, errors.New(pgerrcode.UniqueViolation))
		}
		return insertInfo.ID, nil
	}

}

func (p *PgDB) Set(originalURL string, user string) (*model.StorageInsertInfo, error) {
	var insertInfo model.StorageInsertInfo
	if err := p.sqlInsertURL.Get(&insertInfo, originalURL, user); err != nil {
		return nil, err
	} else {
		return &insertInfo, nil
	}
}

func (p *PgDB) GetURL(id int) (string, error) {
	var originalURL string
	row := p.sqlSelectURL.QueryRowx(id)
	err := row.Scan(&originalURL)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", myerrors.NewNotFoundError()
		}
		return "", err
	}
	return originalURL, nil
}

func (p *PgDB) GetUserURLs(user string) ([]model.StorageURLInfo, error) {
	var URLs []model.StorageURLInfo
	if err := p.sqlSelectUserURLs.Select(&URLs, user); err != nil {
		return nil, err
	}
	if URLs == nil {
		return nil, myerrors.NewNotFoundError()
	}
	return URLs, nil
}
func (p *PgDB) SetBatchURLs(input *[]*model.HandlerURLInfo, username string) (*map[string]*model.StorageURLInfo, error) {
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
