package prototyping

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"

	crud "github.com/go-phings/crud"
	ui "github.com/go-phings/crud-ui"
	sqldb "github.com/go-phings/struct-sql-postgres"
	"github.com/go-phings/umbrella"

	_ "github.com/lib/pq"
)

type Prototype struct {
	dbDSN                   string
	dbTablePrefix           string
	uriAPI                  string
	uriUI                   string
	uriUmbrella             string
	port                    string
	constructors            []func() interface{}
	db                      *sql.DB
	apiCtl                  crud.Controller
	uiCtl                   ui.Controller
	umbrella                umbrella.Umbrella
	umbrellaUserConstructor func() interface{}
	intFieldValues          map[string]ui.IntFieldValues
	stringFieldValues       map[string]ui.StringFieldValues
	orm                     ORM
}

const uriUI = 1
const uriAPI = 2

func (p *Prototype) CreateDB() error {
	db, err := sql.Open("postgres", p.dbDSN)
	if err != nil {
		log.Fatal("Error connecting to db")
	}

	p.orm.SetDatabase(db, p.dbTablePrefix)

	// Append umbrella structs
	if p.umbrellaUserConstructor != nil {
		p.constructors = append(p.constructors, p.umbrellaUserConstructor)
	} else {
		p.constructors = append(p.constructors, func() interface{} { return &umbrella.User{} })
	}
	p.constructors = append(p.constructors, func() interface{} { return &umbrella.Session{} })
	p.constructors = append(p.constructors, func() interface{} { return &umbrella.Permission{} })

	for _, f := range p.constructors {
		o := f()
		err := p.orm.CreateTables(o)
		if err != nil {
			return fmt.Errorf("error with struct db: %w", err)
		}
	}

	noUserConstructor := false
	if p.umbrellaUserConstructor != nil {
		noUserConstructor = true
	}

	p.umbrella = *umbrella.NewUmbrella(db, p.dbTablePrefix, &umbrella.JWTConfig{
		Key:               "protoSecretKey",
		Issuer:            "prototyping.gasior.dev",
		ExpirationMinutes: 5,
	}, &umbrella.UmbrellaConfig{
		TagName:           "ui",
		NoUserConstructor: noUserConstructor,
		ORM:               p.orm,
	})

	if p.umbrellaUserConstructor != nil {
		// This is purelly for user creation
		p.umbrella.Interfaces = &umbrella.Interfaces{
			User: func() umbrella.UserInterface {
				return &defaultUser{
					ctl:         p.orm,
					user:        p.umbrellaUserConstructor().(userInterface),
					constructor: func() userInterface { return p.umbrellaUserConstructor().(userInterface) },
				}
			},
		}
	}

	// create admin user
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

	adminPerm := &umbrella.Permission{
		Flags:   umbrella.FlagTypeAllow,
		ForType: umbrella.ForTypeUser,
		ForItem: 1, // admin's userid
		Ops:     umbrella.OpsCreate | umbrella.OpsRead | umbrella.OpsUpdate | umbrella.OpsDelete | umbrella.OpsList,
		ToType:  "all",
		ToItem:  0,
	}
	errPerm := p.orm.Save(adminPerm)
	if errPerm != nil {
		return fmt.Errorf("error with confirming admin email: %w", errPerm)
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
		p.orm.SetDatabase(db, p.dbTablePrefix)
	}

	noUserConstructor := false
	if p.umbrellaUserConstructor != nil {
		noUserConstructor = true
	}

	p.umbrella = *umbrella.NewUmbrella(p.db, p.dbTablePrefix, &umbrella.JWTConfig{
		Key:               "protoSecretKey",
		Issuer:            "prototyping.gasior.dev",
		ExpirationMinutes: 15,
	}, &umbrella.UmbrellaConfig{
		TagName:           "2db",
		NoUserConstructor: noUserConstructor,
		ORM:               p.orm,
	})
	p.uiCtl = *ui.NewController(p.db, p.dbTablePrefix, &ui.ControllerConfig{
		PasswordGenerator: func(pass string) string {
			passForDB, err := p.umbrella.GeneratePassword(pass)
			if err != nil {
				return ""
			}
			return passForDB
		},
		IntFieldValues:    p.intFieldValues,
		StringFieldValues: p.stringFieldValues,
		ORM:               p.orm,
	})
	p.apiCtl = *crud.NewController(p.db, p.dbTablePrefix, &crud.ControllerConfig{
		PasswordGenerator: func(pass string) string {
			passForDB, err := p.umbrella.GeneratePassword(pass)
			if err != nil {
				return ""
			}
			return passForDB
		},
		ORM: p.orm,
	})

	if p.umbrellaUserConstructor != nil {
		p.umbrella.Interfaces = &umbrella.Interfaces{
			User: func() umbrella.UserInterface {
				return &defaultUser{
					ctl:         p.orm,
					user:        p.umbrellaUserConstructor().(userInterface),
					constructor: func() userInterface { return p.umbrellaUserConstructor().(userInterface) },
				}
			},
		}
	}

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
		uriUI,
		p.uiCtl.Handler(
			p.uriUI,
			p.constructors...,
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
				uriAPI,
				p.apiCtl.Handler(
					fmt.Sprintf("%s%s/", p.uriAPI, s),
					f,
					crud.HandlerOptions{},
				),
				"",
			), umbrella.HandlerConfig{}),
		)
	}

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", p.port), nil))

	return nil
}

func (p *Prototype) wrapHandlerWithUmbrella(uriType int, h http.Handler, redirectNotLogged string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userId := umbrella.GetUserIDFromRequest(r)

		if userId != 0 {
			user := p.umbrella.Interfaces.User()
			found, _ := user.GetByID(userId)
			if found {
				var ctx context.Context
				if uriType == uriUI {
					ctx = context.WithValue(r.Context(), ui.ContextValue("LoggedUserID"), fmt.Sprintf("%d", userId))
					ctx = context.WithValue(ctx, ui.ContextValue("LoggedUserName"), user.GetExtraField("name"))
				} else {
					ctx = context.WithValue(r.Context(), crud.ContextValue("LoggedUserID"), fmt.Sprintf("%d", userId))
					ctx = context.WithValue(ctx, crud.ContextValue("LoggedUserName"), user.GetExtraField("name"))
				}

				for _, o := range []int{umbrella.OpsList, umbrella.OpsRead, umbrella.OpsCreate, umbrella.OpsUpdate, umbrella.OpsDelete} {
					allowedTypes, _ := p.umbrella.GetUserOperationAllowedTypes(userId, o)
					if uriType == uriUI {
						ctx = context.WithValue(ctx, ui.ContextValue(fmt.Sprintf("AllowedTypes_%d", o)), allowedTypes)
					} else {
						ctx = context.WithValue(ctx, crud.ContextValue(fmt.Sprintf("AllowedTypes_%d", o)), allowedTypes)
					}
				}

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
	p.intFieldValues = cfg.IntFieldValues
	p.stringFieldValues = cfg.StringFieldValues

	if cfg.UserConstructor != nil {
		p.umbrellaUserConstructor = cfg.UserConstructor
	}

	if cfg.ORM != nil {
		p.orm = cfg.ORM
	} else {
		p.orm = newWrappedStruct2db("ui")
	}

	return p, nil
}
