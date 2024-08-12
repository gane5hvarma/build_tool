package buildcontextmanager

import (
	"bytes"
	"fmt"
	"os"

	s3buildmanager "github.com/gane5hvarma/build_tool/buildcontextmanager/s3"
)

type Manager interface {
	Upload(buf *bytes.Buffer, key string) (string, error)
}

func Factory(managertype string, bucket string) (Manager, error) {
	switch managertype {
	case "s3":
		config := map[string]string{
			"AWS_ACCESS_KEY_ID":      os.Getenv("AWS_ACCESS_KEY_ID"),
			"AWS_SECERET_ACCESS_KEY": os.Getenv("AWS_SECERET_ACCESS_KEY"),
			"AWS_REGION":             os.Getenv("AWS_REGION"),
			"bucket":                 bucket,
		}
		return s3buildmanager.New(config), nil
	}
	return nil, fmt.Errorf("manager of type %s doesnt exist", managertype)
}
