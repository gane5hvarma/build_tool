package kube

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	batchv1apply "k8s.io/client-go/applyconfigurations/batch/v1"
	corev1 "k8s.io/client-go/applyconfigurations/core/v1"
	corev1apply "k8s.io/client-go/applyconfigurations/core/v1"
	metav1apply "k8s.io/client-go/applyconfigurations/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Client interface {
	ApplySecret(name string, data map[string][]byte, secretType string) error
	ApplyJob(name string, podSpec *corev1apply.PodSpecApplyConfiguration) error
}

type client struct {
	kubeclient *kubernetes.Clientset
	namespace  string
}

func New(namespace string, kubeconfig string) (Client, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}
	kubeclient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &client{kubeclient: kubeclient, namespace: namespace}, nil

}

func getSecretType(secretType string) (v1.SecretType, error) {
	switch secretType {
	case "docker":
		return v1.DockerConfigJsonKey, nil
	case "s3":
		return v1.SecretTypeOpaque, nil

	}
	return "", fmt.Errorf("secret of type %s is not supported ", secretType)
}

func (c *client) ApplySecret(name string, data map[string][]byte, secretType string) error {
	kubeSecretType, err := getSecretType(secretType)
	if err != nil {
		return err
	}
	kind := "Secret"
	version := "v1"
	_, err = c.kubeclient.CoreV1().Secrets(c.namespace).Apply(context.TODO(), &corev1.SecretApplyConfiguration{
		TypeMetaApplyConfiguration: metav1apply.TypeMetaApplyConfiguration{
			Kind:       &kind,
			APIVersion: &version,
		},
		ObjectMetaApplyConfiguration: &metav1apply.ObjectMetaApplyConfiguration{
			Name: &name,
		},
		Type: &kubeSecretType,
		Data: data,
	}, metav1.ApplyOptions{FieldManager: "kinko-field-manager"})
	return err
}

func (c *client) ApplyJob(name string, podSpec *corev1apply.PodSpecApplyConfiguration) error {
	kind := "Job"
	version := "batch/v1"
	_, err := c.kubeclient.BatchV1().Jobs(c.namespace).Apply(context.TODO(), &batchv1apply.JobApplyConfiguration{
		TypeMetaApplyConfiguration: metav1apply.TypeMetaApplyConfiguration{
			Kind:       &kind,
			APIVersion: &version,
		},
		ObjectMetaApplyConfiguration: &metav1apply.ObjectMetaApplyConfiguration{
			Name: &name,
		},
		Spec: &batchv1apply.JobSpecApplyConfiguration{
			Template: &corev1apply.PodTemplateSpecApplyConfiguration{
				Spec: podSpec,
			},
		},
	}, metav1.ApplyOptions{FieldManager: "kinko-field-manager"})
	return err
}
