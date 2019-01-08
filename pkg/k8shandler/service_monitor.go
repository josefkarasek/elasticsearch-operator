package k8shandler

import (
	"context"
	"fmt"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	v1alpha1 "github.com/openshift/elasticsearch-operator/pkg/apis/logging/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CreateOrUpdateServiceMonitors ensures the existence of ServiceMonitors for Elasticsearch cluster
func CreateOrUpdateServiceMonitors(client client.Client, dpl *v1alpha1.Elasticsearch) error {
	serviceMonitorName := fmt.Sprintf("monitor-%s-%s", dpl.Name, "cluster")
	owner := asOwner(dpl)

	labelsWithDefault := appendDefaultLabel(dpl.Name, dpl.Labels)

	elasticsearchScMonitor := createServiceMonitor(serviceMonitorName, dpl.Namespace, labelsWithDefault)
	addOwnerRefToObject(elasticsearchScMonitor, owner)
	err := client.Create(context.TODO(), elasticsearchScMonitor)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure constructing Elasticsearch ServiceMonitor: %v", err)
	}

	// TODO: handle update - use retry.RetryOnConflict

	return nil
}

func createServiceMonitor(serviceMonitorName, namespace string, labels map[string]string) *monitoringv1.ServiceMonitor {
	svcMonitor := serviceMonitor(serviceMonitorName, namespace, labels)
	labelSelector := metav1.LabelSelector{
		MatchLabels: labels,
	}
	tlsConfig := monitoringv1.TLSConfig{
		CAFile: "/etc/prometheus/configmaps/prometheus-serving-certs-ca-bundle/service-ca.crt",
	}
	endpoint := monitoringv1.Endpoint{
		Port:            "restapi",
		Path:            "/_prometheus/metrics",
		Scheme:          "https",
		BearerTokenFile: "/var/run/secrets/kubernetes.io/serviceaccount/token",
		TLSConfig:       &tlsConfig,
	}
	svcMonitor.Spec = monitoringv1.ServiceMonitorSpec{
		JobLabel:  "monitor-elasticsearch",
		Endpoints: []monitoringv1.Endpoint{endpoint},
		Selector:  labelSelector,
	}
	return svcMonitor
}

func serviceMonitor(serviceMonitorName string, namespace string, labels map[string]string) *monitoringv1.ServiceMonitor {
	return &monitoringv1.ServiceMonitor{
		TypeMeta: metav1.TypeMeta{
			Kind:       monitoringv1.ServiceMonitorsKind,
			APIVersion: monitoringv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceMonitorName,
			Namespace: namespace,
			Labels:    labels,
		},
	}
}
