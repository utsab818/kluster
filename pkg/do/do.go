package do

import (
	"context"
	"fmt"
	"strings"

	"github.com/digitalocean/godo"
	"github.com/utsab818/kluster/pkg/apis/utsab.dev/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func Create(c kubernetes.Interface, spec v1alpha1.KlusterSpec) (string, error) {

	token, err := getToken(c, spec.TokenSecret)
	if err != nil {
		return "", nil
	}
	client := godo.NewFromToken(token)

	fmt.Println(client)

	request := &godo.KubernetesClusterCreateRequest{
		Name:        spec.Name,
		RegionSlug:  spec.Region,
		VersionSlug: spec.Version,
		NodePools: []*godo.KubernetesNodePoolCreateRequest{
			&godo.KubernetesNodePoolCreateRequest{
				Size:  spec.NodePools[0].Size,
				Name:  spec.NodePools[0].Name,
				Count: spec.NodePools[0].Count,
			},		
		},
	}

	cluster, _, err := client.Kubernetes.Create(context.Background(), request)
	if err != nil {
		return "", nil
	}

	return cluster.ID, nil
}

// check if the cluter state is running
func ClusterState(c kubernetes.Interface, spec v1alpha1.KlusterSpec, id string) (string, error) {
	token, err := getToken(c, spec.TokenSecret)
	if err != nil {
		return "", nil
	}
	client := godo.NewFromToken(token)

	cluster, _, err := client.Kubernetes.Get(context.Background(), id)
	return string(cluster.Status.State), err

}

func getToken(client kubernetes.Interface, sec string) (string, error) {
	namespace := strings.Split(sec, "/")[0]
	name := strings.Split(sec, "/")[1]

	s, err := client.CoreV1().Secrets(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return "", nil
	}

	return string(s.Data["token"]), nil
}
