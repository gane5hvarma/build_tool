package kaniko

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"github.com/gane5hvarma/build_tool/kube"
	v1 "k8s.io/api/core/v1"
	corev1apply "k8s.io/client-go/applyconfigurations/core/v1"
)

type KanikoSecret interface {
	GetSecretData() map[string][]byte
}

type Kaniko struct {
	kubeclient kube.Client
}

func New(client kube.Client) *Kaniko {
	return &Kaniko{kubeclient: client}
}

type dockerSecret struct {
	username string
	password string
}

type awsSecret struct {
	accessKey string
	secretKey string
	region    string
}

var awsSecretVal *awsSecret
var dockerSecretVal *dockerSecret

const DOCKER_SECRET = "docker"
const AWS_SECRET = "s3"
const KANIKOJOB = "kanikojob"
const DOCKER_SECRET_NAME = "docker-secret"

func (kn *Kaniko) CreateSecrets(buildcontexttype string) error {
	buildContextSecretName := fmt.Sprintf("%s-secret", buildcontexttype)

	secret, err := generateSecret(DOCKER_SECRET)
	if err != nil {
		return err
	}
	err = kn.kubeclient.ApplySecret(DOCKER_SECRET_NAME, secret.GetSecretData(), "docker")
	if err != nil {
		return err
	}

	secret, err = generateSecret(buildcontexttype)
	if err != nil {
		return err
	}

	err = kn.kubeclient.ApplySecret(buildContextSecretName, secret.GetSecretData(), buildcontexttype)
	if err != nil {
		return err
	}
	return nil
}

func (ds *dockerSecret) GetSecretData() map[string][]byte {
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", ds.username, ds.password)))
	server := os.Getenv("DOCKER_SERVER")
	dockerConfig := map[string]map[string]map[string]string{
		"auths": {
			server: {
				"auth": auth,
			},
		},
	}
	dockerConfigJSON, _ := json.Marshal(dockerConfig)
	secret := map[string][]byte{
		v1.DockerConfigJsonKey: dockerConfigJSON,
	}
	return secret
}

func getDockerSecret() *dockerSecret {
	if dockerSecretVal == nil {
		dockerSecretVal = &dockerSecret{
			username: os.Getenv("DOCKER_USERNAME"),
			password: os.Getenv("DOCKER_PASSWORD"),
		}

	}
	return dockerSecretVal

}

func getAWSSecret() *awsSecret {
	if awsSecretVal == nil {
		awsSecretVal = &awsSecret{
			accessKey: os.Getenv("AWS_ACCESS_KEY_ID"),
			secretKey: os.Getenv("AWS_SECERET_ACCESS_KEY"),
			region:    os.Getenv("AWS_REGION"),
		}

	}
	return awsSecretVal
}

func (aws *awsSecret) GetSecretData() map[string][]byte {
	return map[string][]byte{
		"AWS_ACCESS_KEY_ID":     []byte(aws.accessKey),
		"AWS_SECRET_ACCESS_KEY": []byte(aws.secretKey),
	}
}

func generateSecret(secretType string) (KanikoSecret, error) {
	switch secretType {
	case DOCKER_SECRET:
		return getDockerSecret(), nil
	case AWS_SECRET:
		return getAWSSecret(), nil
	}
	return nil, fmt.Errorf("kaniko cant generate secret of type %s", secretType)
}

func getName() *string {
	container := "kaniko-container"
	return &container
}
func getImage() *string {
	image := "gcr.io/kaniko-project/executor:latest"
	return &image
}

func getArgs(imageDestination string, buildContextLocation string) []string {
	return []string{
		fmt.Sprintf("--destination=%s", imageDestination),
		fmt.Sprintf("--context=%s", buildContextLocation),
	}
}

func getEnv(awsSecret string) []corev1apply.EnvVarApplyConfiguration {
	AWS_ACCESS_KEY_ID := "AWS_ACCESS_KEY_ID"
	AWS_REGION := "AWS_REGION"
	val := os.Getenv(AWS_REGION)
	AWS_SECRET_ACCESS_KEY := "AWS_SECRET_ACCESS_KEY"
	return []corev1apply.EnvVarApplyConfiguration{
		{
			Name:  &AWS_REGION,
			Value: &val,
		},
		{
			Name: &AWS_ACCESS_KEY_ID,
			ValueFrom: &corev1apply.EnvVarSourceApplyConfiguration{
				SecretKeyRef: &corev1apply.SecretKeySelectorApplyConfiguration{
					LocalObjectReferenceApplyConfiguration: corev1apply.LocalObjectReferenceApplyConfiguration{
						Name: &awsSecret,
					},
					Key: &AWS_ACCESS_KEY_ID,
				},
			},
		},
		{
			Name: &AWS_SECRET_ACCESS_KEY,
			ValueFrom: &corev1apply.EnvVarSourceApplyConfiguration{
				SecretKeyRef: &corev1apply.SecretKeySelectorApplyConfiguration{
					LocalObjectReferenceApplyConfiguration: corev1apply.LocalObjectReferenceApplyConfiguration{
						Name: &awsSecret,
					},
					Key: &AWS_SECRET_ACCESS_KEY,
				},
			},
		},
	}
}

func getVolumeMounts() []corev1apply.VolumeMountApplyConfiguration {
	dockerSecretName := DOCKER_SECRET_NAME
	dockerMountPath := "/kaniko/.docker"
	return []corev1apply.VolumeMountApplyConfiguration{
		{
			Name:      &dockerSecretName,
			MountPath: &dockerMountPath,
		},
	}
}

func generateJobPodSpec(destination string, buildContextLocation string, buildContextType string) *corev1apply.PodSpecApplyConfiguration {
	secretName := DOCKER_SECRET_NAME
	dockerkey := ".dockerconfigjson"
	path := "config.json"
	restartPolicy := v1.RestartPolicyNever
	podSpec := corev1apply.PodSpecApplyConfiguration{
		Containers: []corev1apply.ContainerApplyConfiguration{
			{
				Name:         getName(),
				Image:        getImage(),
				Args:         getArgs(destination, buildContextLocation),
				Env:          getEnv(fmt.Sprintf("%s-secret", buildContextType)),
				VolumeMounts: getVolumeMounts(),
			},
		},
		RestartPolicy: &restartPolicy,
		Volumes: []corev1apply.VolumeApplyConfiguration{
			{
				Name: &secretName,
				VolumeSourceApplyConfiguration: corev1apply.VolumeSourceApplyConfiguration{
					Secret: &corev1apply.SecretVolumeSourceApplyConfiguration{
						SecretName: &secretName,
						Items: []corev1apply.KeyToPathApplyConfiguration{
							{
								Key:  &dockerkey,
								Path: &path,
							},
						},
					},
				},
			},
		},
	}
	return &podSpec
}

func (kn *Kaniko) CreateJob(jobname string, location string, buildContextType string) error {
	dockerUsername := os.Getenv("DOCKER_USERNAME")
	dockerRepo := os.Getenv("DOCKER_REPO")
	dockerTag := os.Getenv("DOCKER_TAG")
	destination := fmt.Sprintf("%s/%s:%s", dockerUsername, dockerRepo, dockerTag)
	podSpec := generateJobPodSpec(destination, location, buildContextType)
	return kn.kubeclient.ApplyJob(jobname, podSpec)
}
