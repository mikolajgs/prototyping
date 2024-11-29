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
	dbDSN string
	dbTablePrefix string
	uriAPI string
	uriUI string
	port string
	constructors []func() interface{}
	db *sql.DB
	apiCtl restapi.Controller
	uiCtl ui.Controller
}

func (p *Prototype) CreateDB() error {
	db, err := sql.Open("postgres", p.dbDSN)
	if err != nil {
		log.Fatal("Error connecting to db")
	}

	stDB := stdb.NewController(db, p.dbTablePrefix, nil)
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
		db, err := sql.Open("postgres", p.dbDSN)
		if err != nil {
			return errors.New("error connecting to db")
		}
		p.db = db
	}

	p.apiCtl = *restapi.NewController(p.db, p.dbTablePrefix, nil)
	for _, f := range p.constructors {
		s := sqldb.GetStructName(f())

		http.Handle(
			fmt.Sprintf("%s%s", p.uriAPI, s),
			p.apiCtl.Handler(
				fmt.Sprintf("%s%s", p.uriAPI, s), 
				f,
				restapi.HandlerOptions{},
			),
		)
	}

	p.uiCtl = *ui.NewController(p.db, p.dbTablePrefix)
	http.Handle(p.uriUI, p.uiCtl.Handler(
		p.uriUI,
		p.constructors...,
	))

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", p.port), nil))
	return nil
}

func NewPrototype(cfg Config, constructors ...func() interface{}) (*Prototype, error) {
	err := validateConfig(&cfg)
	if err != nil {
		return nil, fmt.Errorf("error with config validation: %w", err)
	}

	p := &Prototype{}
	p.dbDSN = cfg.DatabaseDSN
	p.constructors = constructors
	p.dbTablePrefix = "proto_"
	p.uriAPI = "/api/"
	p.uriUI = "/ui/"
	p.port = "9001"
	return p, nil
}
