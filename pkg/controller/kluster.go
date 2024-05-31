package controller

import (
	"context"
	"log"
	"time"

	"github.com/kanisterio/kanister/pkg/poll"
	"github.com/utsab818/kluster/pkg/apis/utsab.dev/v1alpha1"
	klientset "github.com/utsab818/kluster/pkg/client/clientset/versioned"
	customscheme "github.com/utsab818/kluster/pkg/client/clientset/versioned/scheme"
	kinf "github.com/utsab818/kluster/pkg/client/informers/internalversion/v1alpha1/internalversion"
	klister "github.com/utsab818/kluster/pkg/client/listers/v1alpha1/internalversion"
	"github.com/utsab818/kluster/pkg/do"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
)

type Controller struct {
	client kubernetes.Interface

	// clientset for custom resource kluster
	klient klientset.Interface
	// kluster cache has synced
	klusterSynced cache.InformerSynced
	// lister
	kLister klister.KlusterLister
	// queue
	wq workqueue.RateLimitingInterface
	// Get the event when describing the resource
	recorder record.EventRecorder
}

func NewController(client kubernetes.Interface, klient klientset.Interface, klusterInformer kinf.KlusterInformer) *Controller {
	runtime.Must(customscheme.AddToScheme(scheme.Scheme))

	eveBroadCaster := record.NewBroadcaster()
	eveBroadCaster.StartStructuredLogging(0)
	eveBroadCaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{
		Interface: client.CoreV1().Events(""),
	})
	recorder := eveBroadCaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: "Kluster"})

	c := &Controller{
		client:        client,
		klient:        klient,
		klusterSynced: klusterInformer.Informer().HasSynced,
		kLister:       klusterInformer.Lister(),
		wq:            workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Kluster"),
		recorder:      recorder,
	}

	klusterInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    c.handleAdd,
			DeleteFunc: c.handleDel,
		},
	)

	return c
}

func (c *Controller) Run(ch chan struct{}) error {
	// before running controller we should make sure
	// local cache in informer is initialized at least once.
	if ok := cache.WaitForCacheSync(ch, c.klusterSynced); !ok {
		log.Println("cache was not synced")
	}

	wait.Until(c.worker, time.Second, ch)
	<-ch
	return nil
}

func (c *Controller) worker() {
	for c.processNextItem() {

	}
}

// This function will be called to do things for the kluster
func (c *Controller) processNextItem() bool {
	item, shutDown := c.wq.Get()
	if shutDown {
		return false
	}

	defer c.wq.Forget(item)
	key, err := cache.MetaNamespaceKeyFunc(item)

	if err != nil {
		log.Printf("error %s calling Namespace key func on cache for item", err.Error())
		return false
	}

	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		log.Printf("splitting key into namespace and name, error %s\n", err.Error())
		return false
	}

	kluster, err := c.kLister.Klusters(ns).Get(name)
	if err != nil {
		log.Printf("error %s, Getting the kluster resource from lister", err.Error())
		return false
	}

	log.Printf("kluster spec that we have is %+v\n", kluster.Spec)

	clusterID, err := do.Create(c.client, kluster.Spec)
	if err != nil {
		log.Printf("error %s, creating the cluster", err.Error())
	}

	c.recorder.Event(kluster, corev1.EventTypeNormal, "ClusterCreation", "DO API was called to create the cluster")

	log.Printf("cluster id that we have is %s\n", clusterID)
	err = c.updateStatus(clusterID, "creating", kluster)
	if err != nil {
		log.Printf("error %s, updating status of the kluster %s\n", err.Error(), kluster.Name)
	}

	// wait for digital ocean kubernetes cluster to complete
	// to update the progress from 'creating' to 'Running'
	// query DO API to make sure cluster state is running
	// you can do so manually but can use kanister poll package.
	err = c.WaitForCluster(kluster.Spec, clusterID)
	if err != nil {
		log.Printf("error %s, waiting for cluster to be in running state", err.Error())
	}

	// Now update the status
	err = c.updateStatus(clusterID, "running", kluster)
	if err != nil {
		log.Printf("error %s, updating cluster status after waiting for cluster", err.Error())
	}

	c.recorder.Event(kluster, corev1.EventTypeNormal, "ClusterCreationCompleted", "DO Cluster creation was complete")

	return true
}

func (c *Controller) WaitForCluster(spec v1alpha1.KlusterSpec, clusterId string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	return poll.Wait(ctx, func(ctx context.Context) (bool, error) {
		state, err := do.ClusterState(c.client, spec, clusterId)
		if err != nil {
			return false, err
		}
		if state == "running" {
			return true, nil
		}
		return false, nil
	})
}

func (c *Controller) updateStatus(id, progress string, kluster *v1alpha1.Kluster) error {
	// get the latest version of kluster
	k, err := c.klient.UtsabV1alpha1().Klusters(kluster.Namespace).Get(context.Background(), kluster.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	k.Status.KlusterID = id
	k.Status.Progress = progress
	_, err = c.klient.UtsabV1alpha1().Klusters(kluster.Namespace).UpdateStatus(context.Background(), k, metav1.UpdateOptions{})
	return err
}

func (c *Controller) handleAdd(obj interface{}) {
	log.Println("handleAdd was called")
	c.wq.Add(obj)
}

func (c *Controller) handleDel(obj interface{}) {
	log.Println("handleDel was called")
	c.wq.Add(obj)
}
