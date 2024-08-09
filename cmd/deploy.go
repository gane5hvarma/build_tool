package cmd

import (
	"fmt"

	"os"
	"path/filepath"

	"github.com/gane5hvarma/build_tool/buildcontextmanager"
	"github.com/gane5hvarma/build_tool/compress"
	"github.com/gane5hvarma/build_tool/kaniko"
	"github.com/gane5hvarma/build_tool/kube"
	"github.com/spf13/cobra"
)

var deployCMD = &cobra.Command{
	Use:  "deploy",
	RunE: deploy,
}

var projectDir *string
var buildcontexttype *string
var namespace *string
var kubeconfig *string
var bucket *string

func init() {
	projectDir = deployCMD.Flags().StringP("project-dir", "p", ".", "absolute path to project dir")
	buildcontexttype = deployCMD.Flags().StringP("build-context-manager", "m", "", "build context manager")
	namespace = deployCMD.Flags().StringP("namespace", "n", "default", "namespace to deploy, defaults to default")
	bucket = deployCMD.Flags().StringP("build-context-bucket", "b", "", "bucket for kaniko to fetch build artifacts")
}

var deploy = func(cmd *cobra.Command, args []string) error {
	dockerfilePath := filepath.Join(*projectDir, "Dockerfile")
	if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
		return fmt.Errorf("dockerfile does not exist in the project directory: %s", *projectDir)
	} else if err != nil {
		return fmt.Errorf("error checking Dockerfile: %v", err)
	}
	client, err := kube.New(*namespace, os.Getenv("KUBE_CONFIG"))
	if err != nil {
		return err
	}
	buf, err := compress.CompressDirectory(*projectDir)
	if err != nil {
		return err
	}
	buildContextManager, err := buildcontextmanager.Factory(*buildcontexttype, map[string]string{
		"AWS_ACCESS_KEY_ID":      os.Getenv("AWS_ACCESS_KEY_ID"),
		"AWS_SECERET_ACCESS_KEY": os.Getenv("AWS_SECERET_ACCESS_KEY"),
		"AWS_REGION":             os.Getenv("AWS_REGION"),
		"bucket":                 *bucket,
	})
	if err != nil {
		return err
	}
	key := "random.tar.gz"
	location, err := buildContextManager.Upload(buf, key)
	if err != nil {
		return err
	}
	dockerSecretName := "docker-secret"
	awsSecretName := "aws-secret"
	secret, err := kaniko.GenerateSecret(kaniko.DOCKER_SECRET)
	if err != nil {
		return err
	}
	err = client.ApplySecret(dockerSecretName, secret.GetSecretData(), "docker")
	if err != nil {
		return err
	}

	secret, err = kaniko.GenerateSecret(kaniko.AWS_SECRET)
	if err != nil {
		return err
	}

	err = client.ApplySecret(awsSecretName, secret.GetSecretData(), "s3")
	if err != nil {
		return err
	}
	podSpec := kaniko.GenerateJobPodSpec(fmt.Sprintf("%s/%s:%s", os.Getenv("DOCKER_USERNAME"), os.Getenv("DOCKER_REPO"), "latest"), location)
	if err != nil {
		return err
	}
	err = client.ApplyJob("kanikojob", podSpec)
	if err != nil {
		return err
	}
	return nil
}
