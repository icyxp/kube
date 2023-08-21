package locallb

import (
	"fmt"

	"github.com/icyxp/lvscare/service"
	v1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/staticpod"
)

// LVScare  is
var LVScare Config

// Config is local lb config
type Config struct {
	Masters []string
	Image   string   // default is icyboy/lvscare:latest
	VIP     string   // efault is 10.103.97.2:6443
	Command []string // [lvscare care --vs 192.168.6.10:6443 --rs 127.0.0.1:8081 --rs 127.0.0.1:8082 --rs 127.0.0.1:8083 --health-path /
}

func getLVScarePod() v1.Pod {
	v := make(map[string]v1.Volume)
	t := true
	pod := staticpod.ComponentPod(v1.Container{
		Name:            "kube-lvscare",
		Image:           LVScare.Image,
		ImagePullPolicy: v1.PullIfNotPresent,
		Command:         LVScare.Command,
		SecurityContext: &v1.SecurityContext{Privileged: &t},
	}, v, nil)
	pod.Spec.HostNetwork = true
	return pod
}

// LVScareStaticPodToDisk is
func LVScareStaticPodToDisk(manifests string) {
	staticpod.WriteStaticPodToDisk("kube-lvscare", manifests, getLVScarePod())
}

// InitConfig is
func InitConfig() {
	LVScare.Command = []string{
		"/usr/bin/lvscare",
		"care",
		"--vs",
		LVScare.VIP,
		"--health-path",
		"/healthz",
		"--health-schem",
		"https",
	}

	for _, m := range LVScare.Masters {
		LVScare.Command = append(LVScare.Command, "--rs", m)
	}

	fmt.Printf("lvscare command is: %s\n", LVScare.Command)
}

// CreateLVSFirstTime is
func CreateLVSFirstTime() {
	vs := LVScare.VIP
	rs := LVScare.Masters

	lvs := service.BuildLvscare()

	var errs []string
	//check virturl server
	isAvailable := lvs.IsVirtualServerAvailable(vs)
	if !isAvailable {
		err := lvs.CreateVirtualServer(vs, true)
		if err != nil {
			errs = append(errs, err.Error())
		}
	}

	for _, r := range rs {
		err := lvs.CreateRealServer(r, true)
		if err != nil {
			errs = append(errs, err.Error())
		}
	}

	if len(errs) != 0 {
		fmt.Println("creat ipvs error: ", errs)
	} else {
		fmt.Println("creat ipvs first time", vs, rs)
	}
}

// CreateLocalLB is
func CreateLocalLB() {
	InitConfig()
	CreateLVSFirstTime()
}
