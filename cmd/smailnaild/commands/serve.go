package commands

import (
	"context"
	"errors"
	"net/http"
	"time"

	claysql "github.com/go-go-golems/clay/pkg/sql"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	hostedapp "github.com/go-go-golems/smailnail/pkg/smailnaild"
	"github.com/go-go-golems/smailnail/pkg/smailnaild/accounts"
	hostedauth "github.com/go-go-golems/smailnail/pkg/smailnaild/auth"
	"github.com/go-go-golems/smailnail/pkg/smailnaild/identity"
	"github.com/go-go-golems/smailnail/pkg/smailnaild/rules"
	"github.com/go-go-golems/smailnail/pkg/smailnaild/secrets"
	"github.com/go-go-golems/smailnail/pkg/smailnaild/web"
	"github.com/rs/zerolog/log"
)

type ServeCommand struct {
	*cmds.CommandDescription
}

type ServeSettings struct {
	ListenHost string `glazed:"listen-host"`
	ListenPort int    `glazed:"listen-port"`
}

var _ cmds.BareCommand = &ServeCommand{}

func NewServeCommand() (*ServeCommand, error) {
	defaultSection, err := schema.NewSection(
		schema.DefaultSlug,
		"Hosted Server Settings",
		schema.WithFields(
			fields.New(
				"listen-host",
				fields.TypeString,
				fields.WithHelp("Host interface to bind"),
				fields.WithDefault("0.0.0.0"),
			),
			fields.New(
				"listen-port",
				fields.TypeInteger,
				fields.WithHelp("Port to listen on"),
				fields.WithDefault(8080),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	sqlSection, err := claysql.NewSqlConnectionParameterLayer()
	if err != nil {
		return nil, err
	}

	dbtSection, err := claysql.NewDbtParameterLayer()
	if err != nil {
		return nil, err
	}

	encryptionSection, err := secrets.NewSection()
	if err != nil {
		return nil, err
	}

	authSection, err := hostedauth.NewSection()
	if err != nil {
		return nil, err
	}

	return &ServeCommand{
		CommandDescription: cmds.NewCommandDescription(
			"serve",
			cmds.WithShort("Start the hosted smailnail application backend"),
			cmds.WithLong(`Start the hosted smailnail application backend.

This slice provides:
- Clay SQL-backed application database bootstrapping
- hosted auth mode configuration for dev, session, and future OIDC login
- encrypted IMAP credential storage using the encryption section
- hosted account CRUD, account test, mailbox preview, and rule dry-run APIs
- health, readiness, and service metadata endpoints`),
			cmds.WithSections(defaultSection, sqlSection, dbtSection, encryptionSection, authSection),
		),
	}, nil
}

func (c *ServeCommand) Run(ctx context.Context, parsedValues *values.Values) error {
	settings := &ServeSettings{}
	if err := parsedValues.DecodeSectionInto(schema.DefaultSlug, settings); err != nil {
		return err
	}

	db, dbInfo, err := hostedapp.OpenApplicationDB(ctx, parsedValues)
	if err != nil {
		return err
	}
	defer func() {
		_ = db.Close()
	}()

	if err := hostedapp.BootstrapApplicationDB(ctx, db); err != nil {
		return err
	}

	authSettings, err := hostedauth.LoadSettingsFromParsedValues(parsedValues)
	if err != nil {
		return err
	}

	secretConfig, err := secrets.LoadConfigFromParsedValues(parsedValues)
	if err != nil {
		return err
	}

	accountService := accounts.NewService(accounts.NewRepository(db), secretConfig)
	ruleService := rules.NewService(rules.NewRepository(db), accountService)
	identityRepo := identity.NewRepository(db)
	identityService := identity.NewService(identityRepo)

	userResolver := hostedapp.UserResolver(hostedapp.HeaderUserResolver{
		DefaultUserID: authSettings.DevUserID,
	})
	if authSettings.Mode == hostedauth.AuthModeSession || authSettings.Mode == hostedauth.AuthModeOIDC {
		userResolver = hostedapp.SessionUserResolver{
			Repo:       identityRepo,
			CookieName: authSettings.SessionCookieName,
		}
	}
	var webAuth hostedauth.WebHandler
	if authSettings.Mode == hostedauth.AuthModeOIDC {
		webAuth, err = hostedauth.NewOIDCAuthenticator(ctx, authSettings, identityRepo, identityService)
		if err != nil {
			return err
		}
	}

	server := hostedapp.NewHTTPServer(hostedapp.ServerOptions{
		Host:         settings.ListenHost,
		Port:         settings.ListenPort,
		DB:           db,
		DBInfo:       dbInfo,
		UserResolver: userResolver,
		AccountAPI:   accountService,
		RuleAPI:      ruleService,
		WebAuth:      webAuth,
		PublicFS:     web.PublicFS,
	})

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = hostedapp.ShutdownServer(shutdownCtx, server)
	}()

	log.Info().
		Str("address", server.Addr).
		Str("db_driver", dbInfo.Driver).
		Str("db_target", dbInfo.Target).
		Msg("Starting smailnaild")

	err = server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}
