package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	appconfig "payment-service/app/app_config"
	"payment-service/app/logger"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const DBDriverPostgres = "postgres"

var myApp *application

type application struct {
	id            string
	env           string
	router        *chi.Mux
	httpServer    *httpServer
	debug         *logger.Debug
	db            []dbConfig
	dbConnections map[string]dbConnection
	config        *viper.Viper
	logger        *logger.Logger
}

type dbConfig struct {
	Name             string      `json:"name"`
	Driver           string      `json:"driver"`
	DSN              string      `json:"dsn"`
	Host             string      `json:"host"`
	Port             string      `json:"port"`
	Username         string      `json:"username"`
	Password         string      `json:"password"`
	DbName           string      `json:"dbName"`
	IsCluster        bool        `json:"isCluster"`
	ClusterAddresses []string    `json:"clusterAddresses"`
	Options          interface{} `json:"options"`
}
type dbConnection struct {
	config     dbConfig
	name       string
	connection interface{}
}

type httpServer struct {
	host     string
	port     string
	cors     bool
	httpLogs bool
}

func App() *application {
	if myApp == nil {
		return newApp()
	}
	return myApp
}

func newApp() *application {
	myApp = &application{}
	myApp.configure()
	return myApp
}

func (app *application) Env() string {
	return app.env
}

func (app *application) Id() string {
	return app.id
}
func (app *application) Debug() *logger.Debug {
	return app.debug
}
func (app *application) Router() *chi.Mux {
	return app.router
}

func (app *application) configure() *application {
	app.readConfig().
		setupLogger().
		setConfig().
		setupDbConnections().
		setupRouter()
	return app
}

func (app *application) setupLogger() *application {
	logger := logger.GetLogger()
	app.logger = logger
	return app
}

func (app *application) Logger() *logger.Logger {
	return app.logger
}

func (app *application) setupRouter() *application {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.NoCache)
	r.Use(middleware.Timeout(30 * time.Second))

	l := logrus.New()
	l.Formatter = &logrus.JSONFormatter{
		// disable, as we set our own
		DisableTimestamp: true,
	}
	r.Use(logger.NewStructuredLogger(l))
	app.router = r
	return app

}

func (app *application) setConfig() *application {

	app.Logger().Debug("setting config")

	app.id = app.Config().GetString("app.id")
	app.env = app.Config().GetString("app.env")
	app.httpServer = &httpServer{
		host:     app.Config().GetString("app.host"),
		port:     app.Config().GetString("app.port"),
		cors:     app.Config().GetBool("app.cors"),
		httpLogs: app.Config().GetBool("app.httpLogs"),
	}

	app.debug = app.Logger().Config
	db := app.config.AllSettings()["db"]
	dbConfiguration := map[string]dbConfig{}

	configJson, err := json.Marshal(db)
	if err != nil {
		app.Logger().Error(err)
	}

	fmt.Println(string(configJson))
	err = json.Unmarshal(configJson, &dbConfiguration)
	if err != nil {
		app.Logger().Error(err)
	}

	finalDbConfig := make([]dbConfig, 0, len(dbConfiguration))
	for _, config := range dbConfiguration {
		finalDbConfig = append(finalDbConfig, config)
	}

	app.db = finalDbConfig
	return app
}

func (app *application) Config() *viper.Viper {
	return app.config
}

var readConfig = false

func (app *application) readConfig() *application {
	app.config = appconfig.ReadConfig()
	return app
}

func (app *application) SetRoutes(routes []Route) *application {
	r := app.router
	for _, route := range routes {
		r.MethodFunc(route.Method, route.Pattern, route.HandlerFunc)
	}
	return app

}

func (app *application) StartServer() {

	// print app name
	app.Logger().Infof("Starting project %s", app.id)

	app.Logger().Infof("server started http://%s:%s", app.httpServer.host, app.httpServer.port)
	err := http.ListenAndServe("0.0.0.0:"+app.httpServer.port, app.router)
	if err != nil {
		app.Logger().Panic(err)
	}
}
