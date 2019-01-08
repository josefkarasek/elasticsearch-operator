package k8shandler

import (
	"context"
	"fmt"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/sirupsen/logrus"
)

func createOrUpdatePersistentVolumeClaim(client client.Client, pvc v1.PersistentVolumeClaimSpec, newName string, namespace string) error {
	claim := &v1.PersistentVolumeClaim{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: newName, Namespace: namespace}, claim)
	if err != nil {
		// PVC doesn't exists, needs to be created.
		claim = createPersistentVolumeClaim(newName, namespace, pvc)
		logrus.Infof("Creating new PVC: %v", newName)
		err = client.Create(context.TODO(), claim)
		if err != nil {
			return fmt.Errorf("Unable to create PVC: %v", err)
		}
	} else {
		logrus.Infof("Reusing existing PVC: %s", newName)
		// TODO for updates, don't forget to use retry.RetryOnConflict
	}
	return nil
}

func createPersistentVolumeClaim(pvcName, namespace string, volSpec v1.PersistentVolumeClaimSpec) *v1.PersistentVolumeClaim {
	pvc := persistentVolumeClaim(pvcName, namespace)
	pvc.Spec = volSpec
	return pvc
}

func persistentVolumeClaim(pvcName, namespace string) *v1.PersistentVolumeClaim {
	return &v1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      pvcName,
			Namespace: namespace,
		},
	}
}
