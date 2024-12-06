package prototyping

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"

	restapi "github.com/mikolajgs/prototyping/pkg/rest-api"
	stdb "github.com/mikolajgs/prototyping/pkg/struct-db-postgres"
	sqldb "github.com/mikolajgs/prototyping/pkg/struct-sql-postgres"
	"github.com/mikolajgs/prototyping/pkg/ui"
	"github.com/mikolajgs/prototyping/pkg/umbrella"

	_ "github.com/lib/pq"
)

type Prototype struct {
	dbDSN         string
	dbTablePrefix string
	uriAPI        string
	uriUI         string
	uriUmbrella   string
	port          string
	constructors  []func() interface{}
	db            *sql.DB
	apiCtl        restapi.Controller
	uiCtl         ui.Controller
	umbrella      umbrella.Umbrella
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

	p.umbrella = *umbrella.NewUmbrella(db, p.dbTablePrefix, &umbrella.JWTConfig{
		Key:               "protoSecretKey",
		Issuer:            "prototyping.gasior.dev",
		ExpirationMinutes: 5,
	})
	errUmb := p.umbrella.CreateDBTables()
	if errUmb != nil {
		return fmt.Errorf("error with creating umbrella db: %w", errUmb.Unwrap())
	}

	key, errUmb := p.umbrella.CreateUser("admin@example.com", "admin", map[string]string{
		"Name": "admin",
	})
	if errUmb != nil {
		return fmt.Errorf("error with creating admin: %w", errUmb.Unwrap())
	}
	errUmb = p.umbrella.ConfirmEmail(key)
	if errUmb != nil {
		return fmt.Errorf("error with confirming admin email: %w", errUmb.Unwrap())
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
	p.uiCtl = *ui.NewController(p.db, p.dbTablePrefix)
	p.umbrella = *umbrella.NewUmbrella(p.db, p.dbTablePrefix, &umbrella.JWTConfig{
		Key:               "protoSecretKey",
		Issuer:            "prototyping.gasior.dev",
		ExpirationMinutes: 15,
	})

	// /umbrella/
	http.Handle(p.uriUmbrella, p.umbrella.GetHTTPHandler(p.uriUmbrella))

	// /ui/login/
	http.Handle(fmt.Sprintf("%s%s/", p.uriUI, "login"), p.uiCtl.Handler(
		p.uriUI,
		p.constructors...,
	))

	// /ui/r/login/
	http.Handle(fmt.Sprintf("%s%s/", p.uriUI, "r/login"), p.umbrella.GetLoginHTTPHandler(umbrella.HandlerConfig{
		UseCookie:          "UmbrellaToken",
		CookiePath:         p.uriUI,
		SuccessRedirectURL: "/ui/",
		FailureRedirectURL: "/ui/login/",
	}))

	// /ui/r/logout/
	http.Handle(fmt.Sprintf("%s%s/", p.uriUI, "r/logout"), p.umbrella.GetLogoutHTTPHandler(umbrella.HandlerConfig{
		UseCookie:          "UmbrellaToken",
		CookiePath:         p.uriUI,
		FailureRedirectURL: "/ui/",
		SuccessRedirectURL: "/ui/login/",
	}))

	// /ui/ behind umbrella
	http.Handle(p.uriUI, p.umbrella.GetHTTPHandlerWrapper(p.wrapHandlerWithUmbrella(
		p.uiCtl.Handler(
			p.uriUI,
			append(
				p.constructors,
				func() interface{} { return p.umbrella.Interfaces.User().GetUser() },
				func() interface{} { return p.umbrella.Interfaces.Session().GetSession() },
			)...,
		),
		"/ui/login/",
	), umbrella.HandlerConfig{
		UseCookie: "UmbrellaToken",
	}))

	// /api/ behind umbrella
	for _, f := range p.constructors {
		s := sqldb.GetStructName(f())
		http.Handle(
			fmt.Sprintf("%s%s/", p.uriAPI, s),
			p.umbrella.GetHTTPHandlerWrapper(p.wrapHandlerWithUmbrella(
				p.apiCtl.Handler(
					fmt.Sprintf("%s%s/", p.uriAPI, s),
					f,
					restapi.HandlerOptions{},
				),
				"",
			), umbrella.HandlerConfig{}),
		)
	}

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", p.port), nil))

	return nil
}

func (p *Prototype) wrapHandlerWithUmbrella(h http.Handler, redirectNotLogged string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userId := umbrella.GetUserIDFromRequest(r)

		if userId != 0 {
			user := p.umbrella.Interfaces.User()
			found, _ := user.GetByID(userId)
			if found {
				ctx := context.WithValue(r.Context(), "LoggedUserName", user.GetExtraField("name"))
				req := r.WithContext(ctx)
				h.ServeHTTP(w, req)
				return
			}
		}

		if redirectNotLogged != "" {
			w.Header().Set("Location", redirectNotLogged)
			w.WriteHeader(http.StatusSeeOther)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("NoAccess"))
	})
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
	p.uriUmbrella = "/umbrella/"
	p.port = "9001"
	return p, nil
}
