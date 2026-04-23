package kubernetes

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	k8sclientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	clientset *k8sclientset.Clientset
	namespace string
}

func NewClient(kubeconfigPath, namespace string) (*Client, error) {
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("k8s config: %w", err)
	}

	cs, err := k8sclientset.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("k8s clientset: %w", err)
	}

	return &Client{clientset: cs, namespace: namespace}, nil
}

func (c *Client) CreatePostgresDeployment(ctx context.Context, tenantID, password string) error {
	labels := map[string]string{
		"app":    "tenant-db",
		"tenant": tenantID,
	}

	svcName := tenantID + "-svc"
	stsName := tenantID + "-db"

	// 1. Service
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:   svcName,
			Labels: labels,
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: labels,
			Ports: []corev1.ServicePort{{
				Port:       5432,
				TargetPort: intstr.FromInt(5432),
			}},
		},
	}

	_, err := c.clientset.CoreV1().Services(c.namespace).Create(ctx, svc, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("create service: %w", err)
	}

	// 2. StatefulSet
	replicas := int32(1)
	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:   stsName,
			Labels: labels,
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName: svcName,
			Replicas:    &replicas,
			Selector:    &metav1.LabelSelector{MatchLabels: labels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  "postgres",
						Image: "postgres:15-alpine",
						Ports: []corev1.ContainerPort{{ContainerPort: 5432}},
						Env: []corev1.EnvVar{
							{Name: "POSTGRES_PASSWORD", Value: password},
							{Name: "POSTGRES_USER", Value: "postgres"},
							{Name: "POSTGRES_DB", Value: "postgres"},
						},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      "pgdata",
							MountPath: "/var/lib/postgresql/data",
						}},
					}},
				},
			},
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{{
				ObjectMeta: metav1.ObjectMeta{Name: "pgdata"},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
					Resources: corev1.VolumeResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: resource.MustParse("1Gi"),
						},
					},
				},
			}},
		},
	}

	_, err = c.clientset.AppsV1().StatefulSets(c.namespace).Create(ctx, sts, metav1.CreateOptions{})
	if err != nil {
		// Cleanup the service we just made
		_ = c.clientset.CoreV1().Services(c.namespace).Delete(ctx, svcName, metav1.DeleteOptions{})
		return fmt.Errorf("create statefulset: %w", err)
	}

	return nil
}

// DeletePostgresDeployment tears down the Service + StatefulSet for a tenant.
func (c *Client) DeletePostgresDeployment(ctx context.Context, tenantID string) error {
	svcName := tenantID + "-svc"
	stsName := tenantID + "-db"

	err1 := c.clientset.AppsV1().StatefulSets(c.namespace).Delete(ctx, stsName, metav1.DeleteOptions{})
	err2 := c.clientset.CoreV1().Services(c.namespace).Delete(ctx, svcName, metav1.DeleteOptions{})

	if err1 != nil {
		return fmt.Errorf("delete statefulset: %w", err1)
	}
	if err2 != nil {
		return fmt.Errorf("delete service: %w", err2)
	}
	return nil
}
