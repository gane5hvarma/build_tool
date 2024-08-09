package buildcontextmanager

import (
	"bytes"
	"fmt"

	s3buildmanager "github.com/gane5hvarma/build_tool/buildcontextmanager/s3"
)

type Manager interface {
	Upload(buf *bytes.Buffer, key string) (string, error)
}

func Factory(managertype string, config map[string]string) (Manager, error) {

	switch managertype {
	case "s3":
		return s3buildmanager.New(config), nil
	}
	return nil, fmt.Errorf("manager of type %s doesnt exist", managertype)
}
