  package config
    
    import (
    "fmt"
    "os"
    "path/filepath"
    
    "gopkg.in/yaml.v3"
    )
    
    type Registry struct {
    URL      string `yaml:"url"`
    Insecure bool   `yaml:"insecure"`
  }
    
    type AccessScript struct {
    Type string `yaml:"type"` // e.g., "ssh"
    Host string `yaml:"host"`
    User string `yaml:"user"`
    Cmd  string `yaml:"cmd"`
  }
    
    type AccessControl struct {
    Enable  *AccessScript `yaml:"enable"`
    Disable *AccessScript `yaml:"disable"`
  }
    
    type Config struct {
    AuthStore      string                   `yaml:"auth_store"`
    DefaultTimeout int                      `yaml:"default_timeout_sec"`
    SkopeoPath     string                   `yaml:"skopeo_path"`
    HelmPath       string                   `yaml:"helm_path"`
    OrasPath       string                   `yaml:"oras_path"`
    Registries     map[string]Registry      `yaml:"registries"`
    AccessControl  map[string]AccessControl `yaml:"access_control"`
  }
    
    func Load(path string) (*Config, error) {
    b, err := os.ReadFile(path)
    if err != nil {
  return nil, fmt.Errorf("read config: %w", err)
  }
    var c Config
    if err := yaml.Unmarshal(b, &c); err != nil {
  return nil, fmt.Errorf("parse config: %w", err)
  }
    // sane defaults
    if c.AuthStore == "" {
    c.AuthStore = ".harair/auth.json"
  }
    if c.SkopeoPath == "" {
    c.SkopeoPath = "skopeo"
  }
    if c.HelmPath == "" {
    c.HelmPath = "helm"
  }
    if c.DefaultTimeout == 0 {
    c.DefaultTimeout = 120
  }
    return &c, nil
  }

    func EnsureDir(filePath string) error {
    return os.MkdirAll(filepath.Dir(filePath), 0o700)
  }
