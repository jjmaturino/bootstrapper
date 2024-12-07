package launcher

import (
	"github.com/jjmaturino/pulleytakehome/pkg/bootstrapper/platform"
	"log"
)

var StartVM func(service platform.ApiService, engine platform.Engine, deps ...interface{}) error

func init() {
	StartVM = platform.StartVM
}

func Start(service platform.ApiService, platformType platform.Type, engine platform.Engine, deps ...interface{}) {
	var err error

	switch platformType {
	case "virtual_machine":
		err = StartVM(service, engine, deps...)
		if err != nil {
			log.Printf("failed to start vm: %s", err.Error())
		}

	default:
		log.Printf("unsupported platform type %s", platformType)
	}

}
