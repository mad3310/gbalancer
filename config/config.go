// Copyright 2014. All rights reserved.
// Use of this source code is governed by a GPLv3
// Author: Wenming Zhang <zhgwenming@gmail.com>

package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"io/ioutil"
	
	"github.com/zhgwenming/gbalancer/core/yaml"
)

const (
	DEFAULT_UNIX_SOCKET = "/var/lib/mysql/mysql.sock"
)

func LoadConfig(configFile string) (*Configuration, error) {
	file, err := os.Open(configFile)

	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(file)
	config := &Configuration{
		Service:  "galera",
		Addr:     "127.0.0.1",
		Port:     "3306",
		Timeout:  "5",
	}

	err = decoder.Decode(config)

	// for compatible reason, may remove in the future
	// might be needed by the ipvs engine
	if config.Addr != "" {
		tcpAddr := "tcp://" + config.Addr + ":" + config.Port
		config.AddListen(tcpAddr)
	}

	return config, err
}

type Configuration struct {
	Service    string
	ExtCommand string
	User       string
	Pass       string
	Addr       string
	Port       string
	Timeout    string
	Listen     []string
	Backend    []string
}

func (c *Configuration) ListenInfo() string {
	return fmt.Sprintf("Listen on %v, backend: %v", c.Listen, c.Backend)
}

func (c *Configuration) AddListen(listen string) {
	c.Listen = append(c.Listen, listen)
}

func (c *Configuration) GetListenAddrs() ([]ListenAddr, error) {
	laddrs := make([]ListenAddr, 0, len(c.Listen))
	for _, l := range c.Listen {
		protoAddrParts := strings.SplitN(l, "://", 2)
		if len(protoAddrParts) != 2 {
			err := fmt.Errorf("incorrect listen addr %s", l)
			return laddrs, err
		}

		net, laddr := protoAddrParts[0], protoAddrParts[1]

		var addr ListenAddr
		if net == "unix" {
			// unix://default form
			if laddr == "/" || laddr == "default" {
				laddr = DEFAULT_UNIX_SOCKET
			}
		}

		addr = ListenAddr{net, laddr}

		laddrs = append(laddrs, addr)
	}

	return laddrs, nil
}

//整个config文件对应的结构
type Config struct {
	Addr        string       `yaml:"addr"`
	User        string       `yaml:"user"`
	Password    string       `yaml:"password"`
	LogPath     string       `yaml:"log_path"`
	LogLevel    string       `yaml:"log_level"`
	LogSql      string       `yaml:"log_sql"`
	SlowLogTime int          `yaml:"slow_log_time"`
	AllowIps    string       `yaml:"allow_ips"`
	BlsFile     string       `yaml:"blacklist_sql_file"`
	Charset     string       `yaml:"proxy_charset"`
	Nodes       []NodeConfig `yaml:"nodes"`

	Schema SchemaConfig `yaml:"schema"`
}

//node节点对应的配置
type NodeConfig struct {
	Name             string `yaml:"name"`
	DownAfterNoAlive int    `yaml:"down_after_noalive"`
	MaxConnNum       int    `yaml:"max_conns_limit"`

	User     string `yaml:"user"`
	Password string `yaml:"password"`

	Master string `yaml:"master"`
	Slave  string `yaml:"slave"`
}

//schema对应的结构体
type SchemaConfig struct {
	DB        string        `yaml:"db"`
	Nodes     []string      `yaml:"nodes"`
	Default   string        `yaml:"default"` //default route rule
	ShardRule []ShardConfig `yaml:"shard"`   //route rule
}

//range,hash or date
type ShardConfig struct {
	Table         string   `yaml:"table"`
	Key           string   `yaml:"key"`
	Nodes         []string `yaml:"nodes"`
	Locations     []int    `yaml:"locations"`
	Type          string   `yaml:"type"`
	TableRowLimit int      `yaml:"table_row_limit"`
	DateRange     []string `yaml:"date_range"`
}

func ParseConfigData(data []byte) (*Config, error) {
	var cfg Config
	if err := yaml.Unmarshal([]byte(data), &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func ParseConfigFile(fileName string) (*Config, error) {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	return ParseConfigData(data)
}

func WriteConfigFile(cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	execPath, err := os.Getwd()
	if err != nil {
		return err
	}

	configPath := execPath + "/etc/ks.yaml"
	err = ioutil.WriteFile(configPath, data, 0755)
	if err != nil {
		return err
	}

	return nil
}
