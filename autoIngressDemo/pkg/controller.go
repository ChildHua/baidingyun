package pkg

import (
	"context"
	"log"
	"reflect"
	"time"

	core "k8s.io/api/core/v1"
	networkApi "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	ownerReference "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	informer "k8s.io/client-go/informers/core/v1"
	netinformer "k8s.io/client-go/informers/networking/v1"
	"k8s.io/client-go/kubernetes"
	serviceList "k8s.io/client-go/listers/core/v1"
	ingressList "k8s.io/client-go/listers/networking/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const (
	workNum  = 5
	maxRetry = 10
)

type controller struct {
	clientset     kubernetes.Interface
	ingressLister ingressList.IngressLister
	serviceLister serviceList.ServiceLister
	queue         workqueue.RateLimitingInterface
}

func (c *controller) updateServer(old, newobj interface{}) {
	log.Println("work queue in")
	if !reflect.DeepEqual(old, newobj) {
		c.enqueue(newobj)
	}
}

func (c *controller) addServer(obj interface{}) {
	c.enqueue(obj)
}

func (c *controller) deleteIngress(obj interface{}) {
	ingress := obj.(*networkApi.Ingress)
	ownerReference := ownerReference.GetControllerOf(ingress)
	if ownerReference == nil {
		return
	}
	log.Println("delete ingrss", ownerReference.Kind)
	if ownerReference.Kind != "Service" {
		return
	}
	c.queue.Add(ingress.Namespace + "/" + ingress.Name)
}

func (c *controller) enqueue(obj interface{}) {
	s, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		return
	}
	c.queue.Add(s)
}

func (c *controller) worker() {
	for c.processNextItem() {

	}
}

func (c *controller) processNextItem() bool {
	item, shutdown := c.queue.Get()
	if shutdown {
		return false
	}
	defer c.queue.Done(item)
	key := item.(string)
	log.Println("worke queue out:", key)

	err := c.syncService(key)
	if err != nil {
		c.handlerError(key, err)
	}
	return true
}

func (c *controller) handlerError(key string, err error) {
	if c.queue.NumRequeues(key) > maxRetry {
		// runtime.handlerError(err)
		runtime.HandleError(err)
		c.queue.Forget(key)
		return
	}
	c.queue.AddRateLimited(key)
}

func (c controller) syncService(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	// 删除
	service, err := c.serviceLister.Services(namespace).Get(name)
	if errors.IsNotFound(err) {
		return nil
	}

	if err != nil {
		return err
	}

	// 新增和删除
	_, ok := service.GetAnnotations()["ingress/http"]
	ingress, err := c.ingressLister.Ingresses(namespace).Get(name)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	if ok && errors.IsNotFound(err) {
		// 新建
		ig := c.constructIngress(service)
		_, err := c.clientset.NetworkingV1().Ingresses(namespace).Create(context.TODO(), ig, ownerReference.CreateOptions{})
		if err != nil {
			return err
		}
	} else if !ok && ingress != nil {
		// 删除
		err := c.clientset.NetworkingV1().Ingresses(namespace).Delete(context.TODO(), name, ownerReference.DeleteOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *controller) constructIngress(service *core.Service) *networkApi.Ingress {
	ingress := networkApi.Ingress{}
	ingress.ObjectMeta.OwnerReferences = []ownerReference.OwnerReference{
		*ownerReference.NewControllerRef(service, core.SchemeGroupVersion.WithKind("Service")),
	}
	ingress.Name = service.Name
	ingress.Namespace = service.Namespace
	pathType := networkApi.PathTypePrefix
	icn := "nginx"
	ingress.Spec = networkApi.IngressSpec{
		IngressClassName: &icn,
		Rules: []networkApi.IngressRule{
			{
				Host: "example.com",
				IngressRuleValue: networkApi.IngressRuleValue{
					HTTP: &networkApi.HTTPIngressRuleValue{
						Paths: []networkApi.HTTPIngressPath{
							{
								Path:     "/",
								PathType: &pathType,
								Backend: networkApi.IngressBackend{
									Service: &networkApi.IngressServiceBackend{
										Name: service.Name,
										Port: networkApi.ServiceBackendPort{
											Number: 80,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return &ingress
}

func (c *controller) Run(stopCh <-chan struct{}) {
	for i := 0; i < workNum; i++ {
		go wait.Until(c.worker, time.Minute, stopCh)
	}
	<-stopCh
}

func NewController(clientset kubernetes.Interface, serviceinformer informer.ServiceInformer, ingressInformer netinformer.IngressInformer) controller {
	c := controller{
		clientset:     clientset,
		serviceLister: serviceinformer.Lister(),
		ingressLister: ingressInformer.Lister(),
		queue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "ingressManager"),
	}
	serviceinformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.addServer,
		UpdateFunc: c.updateServer,
	})

	ingressInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: c.deleteIngress,
	})

	return c
}
