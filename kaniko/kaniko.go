package kaniko

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	v1 "k8s.io/api/core/v1"
	corev1apply "k8s.io/client-go/applyconfigurations/core/v1"
)

type KanikoSecret interface {
	GetSecretData() map[string][]byte
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

const DOCKER_SECRET = "DOCKER"
const AWS_SECRET = "AWS"

func (ds *dockerSecret) GetSecretData() map[string][]byte {
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", ds.username, ds.password)))
	server := "https://index.docker.io/v1/" // hardcoded but can be extended different registries
	dockerConfig := map[string]map[string]map[string]string{
		"auths": {
			server: {
				"auth": auth,
			},
		},
	}
	dockerConfigJSON, err := json.Marshal(dockerConfig)
	if err != nil {
		fmt.Printf("Error marshaling Docker config JSON: %v\n", err)
		return nil
	}
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

func GenerateSecret(secretType string) (KanikoSecret, error) {
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

// todo args are hardcoded. have to change
func getArgs(imageDestination string, buildContextLocation string) []string {
	return []string{
		fmt.Sprintf("--destination=%s", imageDestination),
		fmt.Sprintf("--context=%s", buildContextLocation),
	}
}

func getEnv() []corev1apply.EnvVarApplyConfiguration {
	AWS_ACCESS_KEY_ID := "AWS_ACCESS_KEY_ID"
	AWS_REGION := "AWS_REGION"
	val := os.Getenv(AWS_REGION)
	AWS_SECRET_ACCESS_KEY := "AWS_SECRET_ACCESS_KEY"
	awsSecret := "aws-secret"
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
	dockerSecretName := "docker-secret"
	dockerMountPath := "/kaniko/.docker"
	return []corev1apply.VolumeMountApplyConfiguration{
		{
			Name:      &dockerSecretName,
			MountPath: &dockerMountPath,
		},
	}
}

func GenerateJobPodSpec(destination string, buildContextLocation string) *corev1apply.PodSpecApplyConfiguration {
	secretName := "docker-secret"
	dockerkey := ".dockerconfigjson"
	path := "config.json"
	restartPolicy := v1.RestartPolicyNever
	podSpec := corev1apply.PodSpecApplyConfiguration{
		Containers: []corev1apply.ContainerApplyConfiguration{
			{
				Name:         getName(),
				Image:        getImage(),
				Args:         getArgs(destination, buildContextLocation),
				Env:          getEnv(),
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
