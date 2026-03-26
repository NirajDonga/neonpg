package k8s

import (
	"context"
	"fmt"

	"github.com/NirajDonga/dbpods/internal/core"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

type K8sProvisioner struct {
	client    *kubernetes.Clientset
	namespace string
}

func NewProvisioner(client *kubernetes.Clientset, namespace string) core.DBProvisioner {
	return &K8sProvisioner{
		client:    client,
		namespace: namespace,
	}
}

func (k *K8sProvisioner) CreateTenantDatabase(ctx context.Context, tenantID, password string) error {

	serviceBlueprint := k.buildService(tenantID)
	statefulSetBlueprint := k.buildStatefulSet(tenantID, password, serviceBlueprint.Name)

	_, err := k.client.CoreV1().Services(k.namespace).Create(ctx, serviceBlueprint, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create network service: %w", err)
	}

	_, err = k.client.AppsV1().StatefulSets(k.namespace).Create(ctx, statefulSetBlueprint, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create database pod and storage: %w", err)
	}

	return nil
}

func (k *K8sProvisioner) buildService(tenantID string) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:   fmt.Sprintf("%s-db-svc", tenantID),
			Labels: map[string]string{"app": "tenant-db", "tenant": tenantID},
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{Port: 5432, TargetPort: intstr.FromInt(5432)},
			},
			Selector: map[string]string{"app": "tenant-db", "tenant": tenantID},
		},
	}
}

func (k *K8sProvisioner) buildStatefulSet(tenantID, password, serviceName string) *appsv1.StatefulSet {

	postgresContainer := corev1.Container{
		Name:  "postgres",
		Image: "postgres:15-alpine",
		Ports: []corev1.ContainerPort{{ContainerPort: 5432}},
		Env: []corev1.EnvVar{
			{Name: "POSTGRES_PASSWORD", Value: password},
			{Name: "POSTGRES_USER", Value: tenantID},
			{Name: "POSTGRES_DB", Value: fmt.Sprintf("%s_data", tenantID)},
		},
		VolumeMounts: []corev1.VolumeMount{
			{Name: "pgdata", MountPath: "/var/lib/postgresql/data"},
		},
	}

	storageVolume := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{Name: "pgdata"},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("1Gi"),
				},
			},
		},
	}

	replicas := int32(1)
	labels := map[string]string{"app": "tenant-db", "tenant": tenantID}

	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:   fmt.Sprintf("%s-db", tenantID),
			Labels: labels,
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName: serviceName,
			Replicas:    &replicas,
			Selector:    &metav1.LabelSelector{MatchLabels: labels},

			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{postgresContainer},
				},
			},
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{storageVolume},
		},
	}
}
