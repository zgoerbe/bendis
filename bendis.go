package bendis

import (
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/gomodule/redigo/redis"
	"github.com/robfig/cron/v3"
	"github.com/zgoerbe/bendis/cache"
	"github.com/zgoerbe/bendis/filesystems/miniofilesystem"
	"github.com/zgoerbe/bendis/filesystems/s3filesystem"
	"github.com/zgoerbe/bendis/filesystems/sftpfilsystem"
	"github.com/zgoerbe/bendis/filesystems/webdavfilesystem"
	"github.com/zgoerbe/bendis/mailer"
	"log"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/CloudyKit/jet/v6"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/zgoerbe/bendis/render"
	"github.com/zgoerbe/bendis/session"
)

const version = "0.1.0"

var myRedisCache *cache.RedisCache
var myBadgerCache *cache.BadgerCache
var redisPool *redis.Pool
var badgerConn *badger.DB

var maintenanceMode bool

// Bendis is the overall type for the Bendis package. Members that are exported in this type
// are available to any application that uses it.
type Bendis struct {
	AppName       string
	Debug         bool
	Version       string
	ErrorLog      *log.Logger
	InfoLog       *log.Logger
	RootPath      string
	Routes        *chi.Mux
	Render        *render.Render
	Session       *scs.SessionManager
	DB            Database
	JetViews      *jet.Set
	config        config
	EncryptionKey string
	Cache         cache.Cache
	Scheduler     *cron.Cron
	Mail          mailer.Mail
	Server        Server
	FileSystems   map[string]interface{}
	S3            s3filesystem.S3
	SFTP          sftpfilsystem.SFTP
	WebDAV        webdavfilesystem.WebDAV
	Minio         miniofilesystem.Minio
}

type Server struct {
	ServerName string
	Port       string
	Secure     bool
	URL        string
}

type config struct {
	port        string
	renderer    string
	cookie      cookieConfig
	sessionType string
	database    databaseConfig
	redis       RedisConfig
	uploads     uploadConfig
}

type uploadConfig struct {
	allowedMimeTypes []string
	maxUploadSize    int64
}

// New reads the .env file, creates our application config, populates the Bendis type with settings
// based on .env values, and creates necessary folders and files if they don't exist
func (b *Bendis) New(rootPath string) error {
	pathConfig := initPaths{
		rootPath:    rootPath,
		folderNames: []string{"handlers", "migrations", "views", "mail", "data", "public", "tmp", "logs", "middleware", "screenshots"},
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
			Pool:         db,
		}
	}

	if os.Getenv("CACHE") == "redis" || os.Getenv("SESSION_TYPE") == "redis" {
		myRedisCache = b.createClientRedisCache()
		b.Cache = myRedisCache
		redisPool = myRedisCache.Conn
	}

	var scheduler = cron.New()
	b.Scheduler = scheduler

	if os.Getenv("CACHE") == "badger" {
		myBadgerCache = b.createClientBadgerCache()
		b.Cache = myBadgerCache
		badgerConn = myBadgerCache.Conn

		_, err = b.Scheduler.AddFunc("@daily", func() {
			_ = myBadgerCache.Conn.RunValueLogGC(0.7)
		})
		if err != nil {
			return err
		}
	}

	b.InfoLog = infoLog
	b.ErrorLog = errorLog
	b.Debug, _ = strconv.ParseBool(os.Getenv("DEBUG"))
	b.Version = version
	b.RootPath = rootPath
	b.Mail = b.createMailer()
	b.Routes = b.routes().(*chi.Mux)

	// file uploads
	exploded := strings.Split(os.Getenv("ALLOWED_FILETYPES"), ",")
	var mimeTypes []string
	for _, m := range exploded {
		mimeTypes = append(mimeTypes, m)
	}

	var maxUploadSize int64
	if max, err := strconv.Atoi(os.Getenv("MAX_UPLOAD_SIZE")); err != nil {
		maxUploadSize = 10 << 20
	} else {
		maxUploadSize = int64(max)
	}

	b.config = config{
		port:     os.Getenv("PORT"),
		renderer: os.Getenv("RENDERER"),
		cookie: cookieConfig{
			name:     os.Getenv("COOKIE_NAME"),
			lifetime: os.Getenv("COOKIE_LIFETIME"),
			persist:  os.Getenv("COOKIE_PERSISTS"),
			secure:   os.Getenv("COOKIE_SECURE"),
			domain:   os.Getenv("COOKIE_DOMAIN"),
		},
		sessionType: os.Getenv("SESSION_TYPE"),
		database: databaseConfig{
			database: os.Getenv("DATABASE_TYPE"),
			dsn:      b.BuildDSN(),
		},
		redis: RedisConfig{
			host:     os.Getenv("REDIS_HOST"),
			password: os.Getenv("REDIS_PASSWORD"),
			prefix:   os.Getenv("REDIS_PREFIX"),
		},
		uploads: uploadConfig{
			maxUploadSize:    maxUploadSize,
			allowedMimeTypes: mimeTypes,
		},
	}

	secure := true
	if strings.ToLower(os.Getenv("SECURE")) == "false" {
		secure = false
	}

	b.Server = Server{
		ServerName: os.Getenv("SERVER_NAME"),
		Port:       os.Getenv("PORT"),
		Secure:     secure,
		URL:        os.Getenv("APP_URL"),
	}

	// create session
	sess := session.Session{
		CookieLifetime: b.config.cookie.lifetime,
		CookiePersist:  b.config.cookie.persist,
		CookieName:     b.config.cookie.name,
		SessionType:    b.config.sessionType,
		CookieDomain:   b.config.cookie.domain,
	}

	switch b.config.sessionType {
	case "redis":
		sess.RedisPool = myRedisCache.Conn
	case "mysql", "postgres", "mariadb", "postgresql":
		sess.DBPool = b.DB.Pool
	}

	b.Session = sess.InitSession()
	b.EncryptionKey = os.Getenv("KEY")

	if b.Debug {
		var views = jet.NewSet(
			jet.NewOSFileSystemLoader(fmt.Sprintf("%s/views", rootPath)),
			jet.InDevelopmentMode(),
		)
		b.JetViews = views
	} else {
		var views = jet.NewSet(
			jet.NewOSFileSystemLoader(fmt.Sprintf("%s/views", rootPath)),
		)
		b.JetViews = views
	}

	b.createRenderer()
	b.FileSystems = b.createFileSystems()
	go b.Mail.ListenForMail()

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
		Session:  b.Session,
	}
	b.Render = &myRenderer
}

func (b *Bendis) createMailer() mailer.Mail {
	port, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))
	m := mailer.Mail{
		Domain:      os.Getenv("MAIL_DOMAIN"),
		Templates:   b.RootPath + "/mail",
		Host:        os.Getenv("SMTP_HOST"),
		Port:        port,
		Username:    os.Getenv("SMTP_USERNAME"),
		Password:    os.Getenv("SMTP_PASSWORD"),
		Encryption:  os.Getenv("SMTP_ENCRYPTION"),
		FromAddress: os.Getenv("FROM_ADDRESS"),
		FromName:    os.Getenv("FROM_NAME"),
		Jobs:        make(chan mailer.Message, 20),
		Results:     make(chan mailer.Result, 20),
		API:         os.Getenv("MAILER_API"),
		APIKey:      os.Getenv("MAILER_KEY"),
		APIUrl:      os.Getenv("MAILER_URL"),
	}
	return m
}

func (b *Bendis) createClientRedisCache() *cache.RedisCache {
	cacheClient := cache.RedisCache{
		Conn:   b.createRedisPool(),
		Prefix: b.config.redis.prefix,
	}
	return &cacheClient
}

func (b *Bendis) createClientBadgerCache() *cache.BadgerCache {
	cacheClient := cache.BadgerCache{
		Conn: b.createBadgerConn(),
	}
	return &cacheClient
}

func (b *Bendis) createRedisPool() *redis.Pool {
	return &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp",
				b.config.redis.host,
				redis.DialPassword(b.config.redis.password))
		},
		//DialContext: nil,
		TestOnBorrow: func(conn redis.Conn, t time.Time) error {
			_, err := conn.Do("PING")
			return err
		},
		MaxIdle:     50,
		MaxActive:   10000,
		IdleTimeout: 240 * time.Second,
		//Wait:            false,
		//MaxConnLifetime: 0,
	}
}

func (b *Bendis) createBadgerConn() *badger.DB {
	db, err := badger.Open(badger.DefaultOptions(b.RootPath + "/tmp/badger"))
	if err != nil {
		b.InfoLog.Println(err)
		return nil
	}

	return db
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

func (b *Bendis) createFileSystems() map[string]interface{} {
	fileSystems := make(map[string]interface{})

	if os.Getenv("S3_KEY") != "" {
		s3 := s3filesystem.S3{
			Key:      os.Getenv("S3_KEY"),
			Secret:   os.Getenv("S3_SECRET"),
			Region:   os.Getenv("S3_REGION"),
			Endpoint: os.Getenv("S3_ENDPOINT"),
			Bucket:   os.Getenv("S3_BUCKET"),
		}
		fileSystems["S3"] = s3
		b.S3 = s3
	}

	if os.Getenv("MINIO_SECRET") != "" {
		useSSL := false
		if strings.ToLower(os.Getenv("MINIO_USESSL")) == "true" {
			useSSL = true
		}

		minio := miniofilesystem.Minio{
			Endpoint: os.Getenv("MINIO_ENDPOINT"),
			Key:      os.Getenv("MINIO_KEY"),
			Secret:   os.Getenv("MINIO_SECRET"),
			UseSSL:   useSSL,
			Region:   os.Getenv("MINIO_REGION"),
			Bucket:   os.Getenv("MINIO_BUCKET"),
		}
		fileSystems["MINIO"] = minio
		b.Minio = minio
	}

	if os.Getenv("SFTP_HOST") != "" {
		sftp := sftpfilsystem.SFTP{
			Host: os.Getenv("SFTP_HOST"),
			User: os.Getenv("SFTP_USER"),
			Pass: os.Getenv("SFTP_PASS"),
			Port: os.Getenv("SFTP_PORT"),
		}
		fileSystems["SFTP"] = sftp
		b.SFTP = sftp
	}

	if os.Getenv("WEBDAV_HOST") != "" {
		webdav := webdavfilesystem.WebDAV{
			Host: os.Getenv("WEBDAV_HOST"),
			User: os.Getenv("WEBDAV_USER"),
			Pass: os.Getenv("WEBDAV_PASS"),
		}
		fileSystems["WEBDAV"] = webdav
		b.WebDAV = webdav
	}

	return fileSystems
}

type RPCServer struct{}

func (r *RPCServer) MaintenanceMode(inMaintenanceMode bool, resp *string) error {
	if inMaintenanceMode {
		maintenanceMode = true
		*resp = "Server in maintenance mode"
	} else {
		maintenanceMode = false
		*resp = "Server live!"
	}
	return nil
}

func (b *Bendis) listenRPC() {
	// if nothing specified for rpc port, don't start
	if os.Getenv("RPC_PORT") != "" {
		b.InfoLog.Println("Starting RPC server on port", os.Getenv("RPC_PORT"))
		err := rpc.Register(new(RPCServer))
		if err != nil {
			b.ErrorLog.Println(err)
			return
		}
		listen, err := net.Listen("tcp", "127.0.0.1:"+os.Getenv("RPC_PORT"))
		if err != nil {
			b.ErrorLog.Println(err)
			return
		}
		for {
			rpcConn, err := listen.Accept()
			if err != nil {
				continue
			}
			go rpc.ServeConn(rpcConn)
		}
	}
}
