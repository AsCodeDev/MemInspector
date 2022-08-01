package server

import (
	"MemInspector/rpc/tars/CommonInfo"
	_ "embed"
	"fmt"
	"github.com/TarsCloud/TarsGo/tars"
	"os"
)

//go:embed tars.conf
var config []byte

type DeviceInfoImpl struct {
}

func (d *DeviceInfoImpl) GetDevice(info *CommonInfo.DeviceInfo) (int32, error) {
	info.DeviceId = "testDevice"
	return 0, nil
}

func StartListening() {
	commonServer := new(CommonInfo.Deivce)
	deviceImpl := new(DeviceInfoImpl)
	//Tried many ways to init tars without config file, but failed.So embed the config file.
	tars.ServerConfigPath = "./tars.conf"
	if _, err := os.Stat("tars.conf"); os.IsNotExist(err) {
		fmt.Println("tars.conf not found,creating...")
		if err := os.WriteFile("tars.conf", config, 0666); err != nil {
			fmt.Printf("write tars.conf error: %s\n", err.Error())
			os.Exit(-1)
		}
	}
	cfg := tars.GetServerConfig()
	commonServer.AddServant(deviceImpl, cfg.App+"."+cfg.Server+".DeviceInfoObj")
	tars.Run()
}
