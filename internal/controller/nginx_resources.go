package controller

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	webv1alpha1 "github.com/Sherpa99/nginx-opr/api/v1alpha1"
)

func constructDeployment(cr *webv1alpha1.Nginx) *appsv1.Deployment {
	replicas := int32(1)
	if cr.Spec.Replicas != nil {
		replicas = *cr.Spec.Replicas
	}

	image := "nginx:latest"
	if cr.Spec.Image != "" {
		image = cr.Spec.Image
	}

	port := int32(80)
	if cr.Spec.Port != 0 {
		port = cr.Spec.Port
	}

	labels := map[string]string{
		"app":       cr.Name,
		"component": "nginx",
		"tier":      "backend",
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  "nginx",
						Image: image,
						Ports: []corev1.ContainerPort{{
							ContainerPort: port,
						}},
					}},
				},
			},
		},
	}
}

func constructService(cr *webv1alpha1.Nginx) *corev1.Service {
	port := int32(80)
	if cr.Spec.Port != 0 {
		port = cr.Spec.Port
	}

	labels := map[string]string{
		"app":       cr.Name,
		"component": "nginx",
		"tier":      "backend",
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{{
				Port:       port,
				TargetPort: intstr.FromInt(int(port)),
				Protocol:   corev1.ProtocolTCP,
			}},
			Type: corev1.ServiceTypeClusterIP,
		},
	}
}
