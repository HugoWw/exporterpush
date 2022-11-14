package config

import (
	"flag"
	"fmt"
	"github.com/exporterpush/global"
	"github.com/exporterpush/pkg/logger"
	setting2 "github.com/exporterpush/pkg/setting"
	"github.com/natefinch/lumberjack"
	"log"
)

var (
	configPath string
)

func init() {
	err := setupFlag()
	if err != nil {
		log.Fatalf("init.setupFlag err: %v", err)
	}

	// for test
	//configPath = "config/config.yaml"

	err = setupSetting()
	if err != nil {
		log.Fatalf("init.setupSetting err: %v", err)
	}

	err = setupLogger()
	if err != nil {
		log.Fatalf("init.setupLogger err: %v", err)
	}
}

func setupFlag() error {
	flag.StringVar(&configPath, "config", "config/config.yaml", "指定要使用的配置文件路径")
	flag.Parse()

	return nil
}

func setupSetting() error {
	setting, err := setting2.NewSetting(configPath)
	if err != nil {
		return err
	}

	err = setting.ReadSection("global", &global.GlobalSetting)
	if err != nil {
		return err
	}

	if global.GlobalSetting.ScrapeInterval < 15 {
		return fmt.Errorf("The value of global.scrape_interval must be greater than 15 ")
	}

	err = setting.ReadSection("barad", &global.BaradSetting)
	if err != nil {
		return err
	}

	err = setting.ReadSection("prometheus", &global.PrometheusSetting)
	if err != nil {
		return err
	}

	err = setting.ReadSection("pushgateway", &global.PushgatewaySetting)
	if err != nil {
		return err
	}

	if len(global.BaradSetting.StaticConfigs) <= 0 {
		return fmt.Errorf("barad static_configs is nil")
	}

	if len(global.BaradSetting.StaticConfigs[0].Destination) <= 0 {
		return fmt.Errorf("barad destination is nil")
	}

	if len(global.PushgatewaySetting.StaticConfigs) <= 0 {
		return fmt.Errorf("pushgateway static_configs is nil")
	}

	if len(global.PushgatewaySetting.StaticConfigs[0].Destination) <= 0 {
		return fmt.Errorf("pushgateway destination is nil")
	}

	return nil
}

func setupLogger() error {

	global.LogObj = logger.NewLogger(&lumberjack.Logger{
		Filename: global.GlobalSetting.LogSetting.LogSavePath + "/" + global.GlobalSetting.LogSetting.LogFileName +
			global.GlobalSetting.LogSetting.LogFileExt,
		MaxSize:   200,
		MaxAge:    10,
		LocalTime: true,
	}, "", log.LstdFlags).WithCallers(2)

	return nil
}
