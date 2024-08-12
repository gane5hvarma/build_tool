package cmd

import (
	"fmt"

	"os"
	"path/filepath"

	"github.com/gane5hvarma/build_tool/buildcontextmanager"
	"github.com/gane5hvarma/build_tool/compress"
	"github.com/gane5hvarma/build_tool/kaniko"
	"github.com/gane5hvarma/build_tool/kube"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var deployCMD = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy creates a kaniko job in k8s cluster, kaniko then builds the image and upload to dockerhub",
	RunE:  deploy,
}

var projectDir *string
var buildcontexttype *string
var namespace *string
var kubeconfig *string
var bucket *string

func init() {
	projectDir = deployCMD.Flags().StringP("project-dir", "p", ".", "absolute path to project dir")
	buildcontexttype = deployCMD.Flags().StringP("build-context-manager", "m", "", "build context manager for kaniko to download, for ex: s3, gcs, blob")
	namespace = deployCMD.Flags().StringP("namespace", "n", "default", "namespace to deploy kaniko job, defaults to default")
	bucket = deployCMD.Flags().StringP("build-context-bucket", "b", "", "bucket for kaniko to fetch build context")
}

var deploy = func(cmd *cobra.Command, args []string) error {
	// project dir validations
	dockerfilePath := filepath.Join(*projectDir, "Dockerfile")
	if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
		return fmt.Errorf("dockerfile does not exist in the project directory: %s", *projectDir)
	} else if err != nil {
		return fmt.Errorf("error checking Dockerfile: %v", err)
	}
	// see if kube is accessible
	kubeconfig := os.Getenv("KUBE_CONFIG")
	fmt.Println("    ▪ KUBECONFIG=", kubeconfig)
	client, err := kube.New(*namespace, kubeconfig)
	if err != nil {
		fmt.Println("")
		return err
	}
	// compress project directory
	buf, err := compress.CompressDirectory(*projectDir)
	if err != nil {
		return err
	}
	// get build context manager for uploading compress dir to bucket
	bcm, err := buildcontextmanager.Factory(*buildcontexttype, *bucket)
	if err != nil {
		return err
	}
	key := uuid.NewString()
	location, err := bcm.Upload(buf, key)
	if err != nil {
		fmt.Println("Failed to upload build context to manager")
		return err
	}
	fmt.Println("build context uploaded to manager")
	kn := kaniko.New(client)
	err = kn.CreateSecrets(*buildcontexttype)
	if err != nil {
		fmt.Println("Failed to create secrets")
		return err
	}
	fmt.Println("Secrets applied....")
	err = kn.CreateJob(key, location, *buildcontexttype)
	if err != nil {
		fmt.Println("Failed to create job ....")
		return err
	}
	fmt.Println("Job created....")
	return err
}
