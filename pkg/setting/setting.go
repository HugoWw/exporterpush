package setting

import (
	"github.com/spf13/viper"
)

/*
全局配置结构体部分
*/

type GlobalS struct {
	ScrapeInterval    int8             `mapstructure:"scrape_interval"`
	Disk              string           `mapstructure:"disk"`
	NetInterface      string           `mapstructure:"net_interface"`
	ScrapeTargetTypes scrapeTargetType `mapstructure:"scrape_target_types"`
	LogSetting        log              `mapstructure:"log"`
}

type scrapeTargetType struct {
	NodeExporter       string `mapstructure:"node_exporter"`
	ClickhouseExporter string `mapstructure:"clickhouse_exporter"`
}

type log struct {
	LogSavePath string `mapstructure:"log_save_path"`
	LogFileName string `mapstructure:"log_file_name"`
	LogFileExt  string `mapstructure:"log_file_ext"`
}

/*
插件配置部分
*/

type staticConfig struct {
	Destination []string          `mapstructure:"destination"`
	Labels      map[string]string `mapstructure:"labels"`
}

type BaradS struct {
	IsUse         bool           `mapstructure:"is_use"`
	AppId         string         `mapstructure:"app_id"`
	InstanceId    string         `mapstructure:"instance_id"`
	NodeId        string         `mapstructure:"node_id"`
	ProjectId     string         `mapstructure:"project_id"`
	Namespace     string         `mapstructure:"namespace"`
	StaticConfigs []staticConfig `mapstructure:"static_configs"`
}

type PrometheusS struct {
	IsUse         bool           `mapstructure:"is_use"`
	StaticConfigs []staticConfig `mapstructure:"static_configs"`
}

type PushgatewayS struct {
	IsUse         bool           `mapstructure:"is_use"`
	JobnName      string         `mapstructure:"job_name"`
	StaticConfigs []staticConfig `mapstructure:"static_configs"`
}

/*
初始化配置读取
*/

type Setting struct {
	vp *viper.Viper
}

func NewSetting(configs string) (*Setting, error) {
	vp := viper.New()
	vp.SetConfigFile(configs)
	if err := vp.ReadInConfig(); err != nil {
		return nil, err
	}

	return &Setting{vp: vp}, nil
}

func (s *Setting) ReadSection(k string, v interface{}) error {
	err := s.vp.UnmarshalKey(k, v)
	if err != nil {
		return err
	}

	return nil
}
