package controller

import (
	"log"
	"time"

	klientset "github.com/utsab818/kluster/pkg/client/clientset/versioned"
	kinf "github.com/utsab818/kluster/pkg/client/informers/internalversion/v1alpha1/internalversion"
	klister "github.com/utsab818/kluster/pkg/client/listers/v1alpha1/internalversion"
	"github.com/utsab818/kluster/pkg/do"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
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
}

func NewController(client kubernetes.Interface, klient klientset.Interface, klusterInformer kinf.KlusterInformer) *Controller {
	c := &Controller{
		client:        client,
		klient:        klient,
		klusterSynced: klusterInformer.Informer().HasSynced,
		kLister:       klusterInformer.Lister(),
		wq:            workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "kluster"),
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

	clusterId, err := do.Create(c.client, kluster.Spec)
	if err != nil {
		log.Printf("error %s, creating the cluster", err.Error())
	}
	log.Printf("cluster if that we have is %s\n", clusterId)

	return true
}

func (c *Controller) handleAdd(obj interface{}) {
	log.Println("handleAdd was called")
	c.wq.Add(obj)
}

func (c *Controller) handleDel(obj interface{}) {
	log.Println("handleDel was called")
	c.wq.Add(obj)
}
