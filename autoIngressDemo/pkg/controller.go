package pkg

import (
	informer "k8s.io/client-go/informers/core/v1"
	netinformer "k8s.io/client-go/informers/networking/v1"
	"k8s.io/client-go/kubernetes"
	serviceList "k8s.io/client-go/listers/core/v1"
	ingressList "k8s.io/client-go/listers/networking/v1"
	"k8s.io/client-go/tools/cache"
)

type controller struct {
	clientset     kubernetes.Interface
	ingressLister ingressList.IngressLister
	serviceLister serviceList.ServiceLister
}

func (c *controller) updateServer(new, obj interface{}) {

}

func (c *controller) addServer(obj interface{}) {

}

func (c *controller) deleteIngress(obj interface{}) {

}

func (c *controller) Run(stopCh <-chan struct{}) {

}

func NewController(clientset kubernetes.Interface, serviceinformer informer.ServiceInformer, ingressInformer netinformer.IngressInformer) controller {
	c := controller{
		clientset:     clientset,
		serviceLister: serviceinformer.Lister(),
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
