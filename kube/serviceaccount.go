package kube

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ServiceAccountOptions struct {
	Name      string
	Namespace string
}

func CreateOrUpdateServiceAccount(clientset *kubernetes.Clientset, ctx context.Context, opts ServiceAccountOptions) error {
	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      opts.Name,
			Namespace: opts.Namespace,
		},
		ImagePullSecrets: []corev1.LocalObjectReference{
			{
				Name: "docker-" + opts.Name,
			},
		},
	}

	if _, err := clientset.CoreV1().ServiceAccounts(opts.Namespace).Create(ctx, serviceAccount, metav1.CreateOptions{}); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("failed to create serviceaccount resource: %v", err)
		}
		fmt.Println("serviceaccount resource successfully updated")
	} else {
		fmt.Println("serviceaccount resource successfully created")
	}

	return nil
}
