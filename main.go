package prototyping

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	restapi "github.com/mikolajgs/prototyping/pkg/rest-api"
	stdb "github.com/mikolajgs/prototyping/pkg/struct-db-postgres"

	_ "github.com/lib/pq"
)

type Prototype struct {
	dbCfg DBConfig
	db *sql.DB
	constructors map[string]func() interface{}
	apiCtl restapi.Controller
	apiCfg APIConfig
}

func (p *Prototype) CreateDB() error {
	db, err := sql.Open("postgres", fmt.Sprintf("host=localhost user=%s password=%s port=%s dbname=%s sslmode=disable", p.dbCfg.User, p.dbCfg.Pass, p.dbCfg.Port, p.dbCfg.Name))
	if err != nil {
		log.Fatal("Error connecting to db")
	}

	stDB := stdb.NewController(db, p.dbCfg.TablePrefix, nil)
	for _, f := range p.constructors {
		o := f()
		err = stDB.CreateTable(o)
		if err != nil {
			return fmt.Errorf("error with struct db: %w", err)
		}
	}

	return nil
}

func (p *Prototype) Run() error {
	p.apiCtl = *restapi.NewController(p.db, p.dbCfg.TablePrefix, nil)
	for s, f := range p.constructors {
		http.Handle(
			fmt.Sprintf("%s%s", p.apiCfg.URI, s),
			p.apiCtl.Handler(
				fmt.Sprintf("%s%s", p.apiCfg.URI, s), 
				f,
				restapi.HandlerOptions{},
			),
		)
	}
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", p.apiCfg.Port), nil))
	return nil
}

func NewPrototype(dbCfg DBConfig, constructors map[string]func() interface{}, apiCfg APIConfig) (*Prototype, error) {
	err := validateDBConfig(&dbCfg)
	if err != nil {
		return nil, fmt.Errorf("init error: %w", err)
	}

	err = validateAPIConfig(&apiCfg)
	if err != nil {
		return nil, fmt.Errorf("init error: %w", err)
	}

	err = validateConstructors(&constructors)
	if err != nil {
		return nil, fmt.Errorf("init error: %w", err)
	}

	p := &Prototype{}
	p.dbCfg = dbCfg
	p.apiCfg = apiCfg
	p.constructors = constructors
	return p, nil
}
