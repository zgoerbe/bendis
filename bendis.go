package bendis

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/CloudyKit/jet/v6"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/zgoerbe/bendis/render"
	"github.com/zgoerbe/bendis/session"
)

const version = "1.0.0"

// Bendis is the overal type for the Bendis package. Members that are exported in this type
// are available to any application that uses it.
type Bendis struct {
	AppName  	string
	Debug    	bool
	Version  	string
	ErrorLog 	*log.Logger
	InfoLog  	*log.Logger
	RootPath 	string
	Routes   	*chi.Mux
	Render   	*render.Render
	Session 	*scs.SessionManager
	DB 			Database
	JetViews 	*jet.Set
	config  	config
}

type config struct {
	port     	string
	renderer 	string
	cookie 		cookieConfig
	sessionType string
	database 	databaseConfig
}

// New reads the .env file, creates our application config, populates the Bendis type with settings
// based on .env values, and creates necessary folders and files if they don't exist
func (b *Bendis) New(rootPath string) error {
	pathConfig := initPaths{
		rootPath:    rootPath,
		folderNames: []string{"handlers", "migrations", "views", "data", "public", "tmp", "logs", "middleware"},
	}

	err := b.Init(pathConfig)
	if err != nil {
		return err
	}

	err = b.checkDotEnv(rootPath)
	if err != nil {
		return err
	}

	// read .env
	err = godotenv.Load(rootPath + "/.env")
	if err != nil {
		return err
	}

	// create loggers
	infoLog, errorLog := b.startLoggers()

	// connect to database
	if os.Getenv("DATABASE_TYPE") != "" {
		db, err := b.OpenDB(os.Getenv("DATABASE_TYPE"), b.BuildDSN())
		if err != nil {
			errorLog.Println(err)
			os.Exit(1)
		}
		b.DB = Database{
			DatabaseType: os.Getenv("DATABASE_TYPE"),
			Pool: db,
		}
	}

	b.InfoLog = infoLog
	b.ErrorLog = errorLog
	b.Debug, _ = strconv.ParseBool(os.Getenv("DEBUG"))
	b.Version = version
	b.RootPath = rootPath
	b.Routes = b.routes().(*chi.Mux)

	b.config = config{
		port:     os.Getenv("PORT"),
		renderer: os.Getenv("RENDERER"),
		cookie: cookieConfig{
			name: os.Getenv("COOKIE_NAME"),
			lifetime: os.Getenv("COOKIE_LIFETIME"),
			persist: os.Getenv("COOKIE_PERSISTS"),
			secure: os.Getenv("COOKIE_SECURE"),
			domain: os.Getenv("COOKIE_DOMAIN"),
		},
		sessionType: os.Getenv("SESSION_TYPE"),
		database: databaseConfig{
			database: os.Getenv("DATABASE_TYPE"),
			dsn: b.BuildDSN(),
		},
	}

	// Create session
	sess := session.Session {
		CookieLifetime: b.config.cookie.lifetime,
		CookiePersist: b.config.cookie.persist,
		CookieName: b.config.cookie.name,
		SessionType: b.config.sessionType,
		CookieDomain: b.config.cookie.domain,
	}

	b.Session = sess.InitSession()

	var views = jet.NewSet(
		jet.NewOSFileSystemLoader(fmt.Sprintf("%s/views", rootPath)),
		jet.InDevelopmentMode(),
	)

	b.JetViews = views

	b.createRenderer()

	return nil
}

// Init creates necessary folders for our Bendis application
func (b *Bendis) Init(p initPaths) error {
	root := p.rootPath
	for _, path := range p.folderNames {
		// create folder if it does not exist
		err := b.CreateDirIfNotExist(root + "/" + path)
		if err != nil {
			return err
		}
	}

	return nil
}

// ListenAndServer starts the web server
func (b *Bendis) ListenAndServer() {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", os.Getenv("PORT")),
		ErrorLog:     b.ErrorLog,
		Handler:      b.Routes,
		IdleTimeout:  30 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 600 * time.Second,
	}

	defer b.DB.Pool.Close()

	b.InfoLog.Printf("Listening on port %s", os.Getenv("PORT"))
	err := srv.ListenAndServe()
	b.ErrorLog.Fatal(err)
}

func (b *Bendis) checkDotEnv(path string) error {
	err := b.CreateFileIfNotExist(fmt.Sprintf("%s/.env", path))
	if err != nil {
		return err
	}

	return nil
}

func (b *Bendis) startLoggers() (*log.Logger, *log.Logger) {
	var infoLog *log.Logger
	var errorLog *log.Logger

	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	return infoLog, errorLog
}

func (b *Bendis) createRenderer() {
	myRenderer := render.Render{
		Renderer: b.config.renderer,
		RootPath: b.RootPath,
		//Secure:     false,
		Port: b.config.port,
		//ServerName: "",
		JetViews: b.JetViews,
		Session: b.Session,
	}
	b.Render = &myRenderer
}

func (b *Bendis) BuildDSN() string {
	var dsn string

	switch os.Getenv("DATABASE_TYPE") {
	case "postgres", "postgresql":
		dsn = fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s timezone=UTC connect_timeout=5",
		os.Getenv("DATABASE_HOST"),
		os.Getenv("DATABASE_PORT"),
		os.Getenv("DATABASE_USER"),
		os.Getenv("DATABASE_NAME"),
		os.Getenv("DATABASE_SSL_MODE"))

		if os.Getenv("DATABASE_PASS") != "" {
			dsn = fmt.Sprintf("%s password=%s", dsn, os.Getenv("DATABASE_PASS"))
		}

	default:

	}

	return dsn
}