package main

import (
	"fmt"
	"learBaiding/autoIngressDemo/pkg"
	"log"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	// build config
	// create clientset
	// create shareinformer
	// create workequeue
	// add event
	// start informer
	c, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		inClusterConfig, err := rest.InClusterConfig()
		if err != nil {
			log.Fatalln(err)
		}
		c = inClusterConfig
	}
	clientset, err := kubernetes.NewForConfig(c)
	if err != nil {
		fmt.Println(err)
		return
	}
	sif := informers.NewSharedInformerFactory(clientset, 0)
	seveiceInformer := sif.Core().V1().Services()
	ingressInformer := sif.Networking().V1().Ingresses()
	ctl := pkg.NewController(clientset, seveiceInformer, ingressInformer)

	stopCh := make(chan struct{})
	sif.Start(stopCh)
	sif.WaitForCacheSync(stopCh)
	ctl.Run(stopCh)
}
