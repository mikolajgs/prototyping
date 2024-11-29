package prototyping

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"

	restapi "github.com/mikolajgs/prototyping/pkg/rest-api"
	stdb "github.com/mikolajgs/prototyping/pkg/struct-db-postgres"
	sqldb "github.com/mikolajgs/prototyping/pkg/struct-sql-postgres"
	"github.com/mikolajgs/prototyping/pkg/ui"

	_ "github.com/lib/pq"
)

type Prototype struct {
	dbCfg DbConfig
	db *sql.DB
	constructors []func() interface{}
	apiCtl restapi.Controller
	uiCtl ui.Controller
	httpCfg HttpConfig
}

func (p *Prototype) CreateDB() error {
	db, err := sql.Open("postgres", fmt.Sprintf("host=localhost user=%s password=%s port=%s dbname=%s sslmode=disable", p.dbCfg.User, p.dbCfg.Pass, p.dbCfg.Port, p.dbCfg.Name))
	if err != nil {
		log.Fatal("Error connecting to db")
	}

	stDB := stdb.NewController(db, p.dbCfg.TablePrefix, nil)
	for _, f := range p.constructors {
		o := f()
		err := stDB.CreateTable(o)
		if err != nil {
			return fmt.Errorf("error with struct db: %w", err.Unwrap())
		}
	}

	db.Close()

	return nil
}

func (p *Prototype) Run() error {
	if p.db == nil {
		db, err := sql.Open("postgres", fmt.Sprintf("host=localhost user=%s password=%s port=%s dbname=%s sslmode=disable", p.dbCfg.User, p.dbCfg.Pass, p.dbCfg.Port, p.dbCfg.Name))
		if err != nil {
			return errors.New("error connecting to db")
		}
		p.db = db
	}

	p.apiCtl = *restapi.NewController(p.db, p.dbCfg.TablePrefix, nil)
	for _, f := range p.constructors {
		s := sqldb.GetStructName(f())

		http.Handle(
			fmt.Sprintf("%s%s", p.httpCfg.ApiUri, s),
			p.apiCtl.Handler(
				fmt.Sprintf("%s%s", p.httpCfg.ApiUri, s), 
				f,
				restapi.HandlerOptions{},
			),
		)
	}

	p.uiCtl = *ui.NewController(p.db, p.dbCfg.TablePrefix)
	http.Handle(p.httpCfg.UiUri, p.uiCtl.Handler(
		p.httpCfg.UiUri,
		p.constructors...,
	))

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", p.httpCfg.Port), nil))
	return nil
}

func NewPrototype(dbCfg DbConfig, constructors []func() interface{}, httpCfg HttpConfig) (*Prototype, error) {
	err := validateDbConfig(&dbCfg)
	if err != nil {
		return nil, fmt.Errorf("init error: %w", err)
	}

	err = validateHttpConfig(&httpCfg)
	if err != nil {
		return nil, fmt.Errorf("init error: %w", err)
	}

	p := &Prototype{}
	p.dbCfg = dbCfg
	p.httpCfg = httpCfg
	p.constructors = constructors
	return p, nil
}
