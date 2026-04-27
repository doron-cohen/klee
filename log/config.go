package log

// RotationConfig controls log file rotation.
type RotationConfig struct {
	MaxSizeMB  int `yaml:"max_size"    env:"LOG_FILE_ROTATION_MAX_SIZE"    default:"100"`
	MaxBackups int `yaml:"max_backups" env:"LOG_FILE_ROTATION_MAX_BACKUPS"  default:"3"`
	MaxAgeDays int `yaml:"max_age"     env:"LOG_FILE_ROTATION_MAX_AGE"      default:"28"`
}

// ConsoleConfig controls the console sink.
type ConsoleConfig struct {
	Enabled bool   `yaml:"enabled" env:"LOG_CONSOLE_ENABLED" default:"true"`
	Level   string `yaml:"level"   env:"LOG_CONSOLE_LEVEL"   default:"info"`
	Format  string `yaml:"format"  env:"LOG_CONSOLE_FORMAT"  default:"pretty"`
}

// FileConfig controls the file sink.
type FileConfig struct {
	Enabled  bool           `yaml:"enabled" env:"LOG_FILE_ENABLED" default:"false"`
	Path     string         `yaml:"path"    env:"LOG_FILE_PATH"`
	Level    string         `yaml:"level"   env:"LOG_FILE_LEVEL"   default:"debug"`
	Format   string         `yaml:"format"  env:"LOG_FILE_FORMAT"  default:"json"`
	Rotation RotationConfig `yaml:"rotation"`
}

// Config is the embeddable log configuration block.
//
// Embed in your app config to get automatic log wiring:
//
//	type AppConfig struct {
//	    log.Config `yaml:"log"`
//	}
type Config struct {
	Console ConsoleConfig `yaml:"console"`
	File    FileConfig    `yaml:"file"`
}

// LogConfig satisfies Provider, enabling auto-wiring when Config is embedded.
func (c Config) LogConfig() Config { return c }
