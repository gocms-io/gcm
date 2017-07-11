package config

import "github.com/gocms-io/gcm/config/config_os"

// global flags
const FLAG_VERBOSE = "verbose"
const FLAG_SET_VERSION = "useVersion"

// binary items
const BINARY_PROTOCOL = "http"
const BINARY_HOST = "release"
const BINARY_DOMAIN = "gocms.io"
const BINARY_OS_PATH = config_os.BINARY_OS_PATH
const BINARY_ARCHIVE = "gocms.zip"
const BINARY_FILE = config_os.BINARY_FILE
const BINARY_DEFAULT_RELEASE = "alpha-release"
const BINARY_DEFAULT_VERSION = "current"

// other dirs and files
const CONTENT_DIR = "content"
const ENV_FILE = ".env"
const DOCS_DIR = "docs"
const PLUGINS_DIR = "plugins"
const TEMPLATES_DIR = "templates"
const THEMES_DIR = "themes"
const THEMES_DEFAULT_DIR = "default"
const GOCMS_ADMIN_DIR = "gocms"
const BACKUP_DIR = ".bk"
const STAGING_DIR = ".staging"
const PLUGIN_MANIFEST = "manifest.json"
