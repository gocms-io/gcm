package plugin_manifest

import (
	"github.com/urfave/cli"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/joho/godotenv/autoload"
	"os"
	"database/sql"
	"github.com/gocms-io/gocms/utility/gocms_plugin_util/manifest_utl"
	"github.com/gocms-io/gocms/utility/errors"
	"fmt"
)

const flag_database = "d"
const flag_database_long = "database"
const flag_user = "u"
const flag_user_long = "user"
const flag_password = "p"
const flag_password_long = "password"
const flag_server = "s"
const flag_server_long = "server"


type pluginManifestContext struct {
	manifestPath string
	dbName string
	dbUser string
	dbPassword string
	dbServer string
}

var CMD_PLUGIN_MANIFEST = cli.Command{
	Name:      "manifest",
	Usage:     "update plugin's manifest in current gocms database according to the file given. Database connection string generated from information in .env",
	ArgsUsage: "<manifest file>",
	Action:    cmd_update_manifest,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  flag_database + ", " + flag_database_long,
			Usage: "database name to use in connection. Defaults to pulling form .env file",
		},
		cli.StringFlag{
			Name:  flag_user + ", " + flag_user_long,
			Usage: "database user to use in connection. Defaults to pulling form .env file",
		},
		cli.StringFlag{
			Name:  flag_password + ", " + flag_password_long,
			Usage: "database password to use in connection. Defaults to pulling form .env file",
		},
		cli.StringFlag{
			Name:  flag_server + ", " + flag_server_long,
			Usage: "database server fqdn to use in connection. Defaults to pulling form .env file",
		},
	},
}

func cmd_update_manifest(c *cli.Context) error {

	// get command context from cli
	pctx, err := buildContextFromFlags(c)
	if err != nil {
		return err
	}

	// get open database connection
	db, err := createMySqlConnectionOrFail(pctx)
	if err != nil {
		return err
	}

	// insert manifest
	err = manifest_utl.InsertManifest(pctx.manifestPath, *db)
	if err != nil {
		return err
	}

	return nil
}


func buildContextFromFlags(c *cli.Context) (*pluginManifestContext, error) {

	pctx := pluginManifestContext{
		dbName:     GetEnvVar("DB_NAME"),
		dbUser:     GetEnvVar("DB_USER"),
		dbPassword: GetEnvVar("DB_PASSWORD"),
		dbServer:   GetEnvVar("DB_SERVER"),
	}

	// dbName
	if c.String(flag_database) != ""{
		pctx.dbName = c.String(flag_database)
	}
	if pctx.dbName == "" {
		err := errors.New("database must be specified. It was not found in the environment vars, .env file, or cli flags.\n")
		fmt.Println(err)
		return nil, err
	}

	// dbUser
	if c.String(flag_user) != ""{
		pctx.dbUser = c.String(flag_user)
	}
	if pctx.dbUser == "" {
		err := errors.New("name must be specified. It was not found in the environment vars, .env file, or cli flags.\n")
		fmt.Println(err)
		return nil, err
	}

	// dbPassword
	if c.String(flag_password) != ""{
		pctx.dbPassword = c.String(flag_password)
	}
	if pctx.dbPassword == "" {
		err := errors.New("password must be specified. It was not found in the environment vars, .env file, or cli flags.\n")
		fmt.Println(err)
		return nil, err
	}

	// dbServer
	if c.String(flag_server) != ""{
		pctx.dbServer = c.String(flag_server)
	}
	if pctx.dbServer == "" {
		err := errors.New("server must be specified. It was not found in the environment vars, .env file, or cli flags.\n")
		fmt.Println(err)
		return nil, err
	}


	manifestPath := c.Args().Get(0)
	// verify src and dest exist
	if manifestPath == "" {
		err := errors.New("A manifest file must be specified")
		fmt.Println(err)
		return nil, err
	}
	pctx.manifestPath = manifestPath



	return &pctx, nil
}


func createMySqlConnectionOrFail(pctx * pluginManifestContext) (*sql.DB, error) {
		// create db connection
		connectionString := pctx.dbName + ":" + pctx.dbPassword + "@" + pctx.dbServer + "/" + pctx.dbName + "?parseTime=true"
		dbHandle, err := sql.Open("mysql", connectionString)
		if err != nil {
			err := errors.New(fmt.Sprintf("Database Error opening connection: %v\n", err.Error()))
			fmt.Println(err)
			return nil, err
	}

		// ping to verify connection
		err = dbHandle.Ping()
		if err != nil {
			err := errors.New(fmt.Sprintf("Database Error verifying good connection: %v\n", err.Error()))
			fmt.Println(err)
			return nil, err
	}

		return dbHandle, nil
}

func GetEnvVar(envVar string) string {
	is := os.Getenv(envVar)
	return is
}