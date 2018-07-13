package stub

import (
	"context"
	"fmt"
	"reflect"

	v1beta1 "github.com/operator-framework/operator-sdk-samples/operator-ipfs/pkg/apis/app/v1beta1"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func NewHandler() sdk.Handler {
	return &Handler{}
}

type Handler struct {
}

func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	switch o := event.Object.(type) {
	case *v1beta1.Deployment:
		ipfs := o

		// Ignore the delete event since the garbage collector will clean up all secondary resources for the CR
		// All secondary resources must have the CR set as their OwnerReference for this to be the case
		if event.Deleted {
			return nil
		}

		// Create the deployment if it doesn't exist
		dep := deploymentForBootstrap(ipfs)
		err := sdk.Create(dep)
		if err != nil && !apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("failed to create deployment: %v", err)
		}

		dep = deploymentForIpfs(ipfs)
		err = sdk.Create(dep)
		if err != nil && !apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("failed to create deployment: %v", err)
		}

		// Ensure the deployment size is the same as the spec
		err = sdk.Get(dep)
		if err != nil {
			return fmt.Errorf("failed to get deployment: %v", err)
		}
		size := ipfs.Spec.Size
		if *dep.Spec.Replicas != size {
			dep.Spec.Replicas = &size
			err = sdk.Update(dep)
			if err != nil {
				return fmt.Errorf("failed to update deployment: %v", err)
			}
		}

		// Update the Ipfs status with the pod names
		podList := podList()
		labelSelector := labels.SelectorFromSet(labelsForIpfs(ipfs.Name)).String()
		listOps := &metav1.ListOptions{LabelSelector: labelSelector}
		err = sdk.List(ipfs.Namespace, podList, sdk.WithListOptions(listOps))
		if err != nil {
			return fmt.Errorf("failed to list pods: %v", err)
		}
		podNames := getPodNames(podList.Items)
		if !reflect.DeepEqual(podNames, ipfs.Status.Nodes) {
			ipfs.Status.Nodes = podNames
			err := sdk.Update(ipfs)
			if err != nil {
				return fmt.Errorf("failed to update ipfs status: %v", err)
			}
		}
	}
	return nil
}

// deploymentForIpfs returns a ipfs Deployment object
func deploymentForIpfs(m *v1beta1.Deployment) *appsv1.Deployment {
	ls := labelsForIpfs(m.Name)
	replicas := m.Spec.Size

	dep := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "extensions/v1beta1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{{
						Image:   "ipfs:1.4.36-alpine",
						Name:    "ipfs",
						Command: []string{`["/usr/local/bin/start-daemons.sh"]`},
						Ports: []v1.ContainerPort{
							{
								ContainerPort: 4001,
								Name:          "swarm",
								Protocol:      "TCP"},
							{
								ContainerPort: 5001,
								Name:          "api",
								Protocol:      "TCP",
							},
							{
								ContainerPort: 9094,
								Name:          "clusterapi",
								Protocol:      "TCP",
							},
							{
								ContainerPort: 9095,
								Name:          "clusterproxy",
								Protocol:      "TCP",
							},
							{
								ContainerPort: 9096,
								Name:          "cluster",
								Protocol:      "TCP",
							},
						},
					}},
				},
			},
		},
	}
	addOwnerRefToObject(dep, asOwner(m))
	return dep
}

func deploymentForBootstrap(m *v1beta1.Deployment) *appsv1.Deployment {
	ls := map[string]string{"app": "ipfs", "role": "bootstrap"}
	replicas := m.Spec.Size

	dep := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "extensions/v1beta1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{{
						Image:   "ipfs:1.4.36-alpine",
						Name:    "ipfs",
						Command: []string{`["/usr/local/bin/start-daemons.sh"]`},
						Ports: []v1.ContainerPort{
							{
								ContainerPort: 4001,
								Name:          "swarm",
								Protocol:      "TCP"},
							{
								ContainerPort: 5001,
								Name:          "api",
								Protocol:      "TCP",
							},
							{
								ContainerPort: 9094,
								Name:          "clusterapi",
								Protocol:      "TCP",
							},
							{
								ContainerPort: 9095,
								Name:          "clusterproxy",
								Protocol:      "TCP",
							},
							{
								ContainerPort: 9096,
								Name:          "cluster",
								Protocol:      "TCP",
							},
						},
					}},
				},
			},
		},
	}
	addOwnerRefToObject(dep, asOwner(m))
	return dep
}

// labelsForIpfs returns the labels for selecting the resources
// belonging to the given ipfs CR name.
func labelsForIpfs(name string) map[string]string {
	return map[string]string{"app": "ipfs", "role": "peer"}
}

// addOwnerRefToObject appends the desired OwnerReference to the object
func addOwnerRefToObject(obj metav1.Object, ownerRef metav1.OwnerReference) {
	obj.SetOwnerReferences(append(obj.GetOwnerReferences(), ownerRef))
}

// asOwner returns an OwnerReference set as the ipfs CR
func asOwner(m *v1beta1.Deployment) metav1.OwnerReference {
	trueVar := true
	return metav1.OwnerReference{
		APIVersion: m.APIVersion,
		Kind:       m.Kind,
		Name:       m.Name,
		UID:        m.UID,
		Controller: &trueVar,
	}
}

// podList returns a v1.PodList object
func podList() *v1.PodList {
	return &v1.PodList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
	}
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNames(pods []v1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}
