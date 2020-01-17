package kube

import "deployfromgo/src/config"

type InitCluster struct {
	ApiVersion           string            `yaml:"apiVersion"`
	Kind                 string            `yaml:"kind"`
	Etcd                 Etcd              `yaml:"etcd"`
	Networking           map[string]string `yaml:"networking"`
	Dns                  map[string]string `yaml:"dns"`
	ControlPlaneEndpoint string            `controlPlaneEndpoint`
	ApiServer            Apiserver         `yaml:"apiServer"`
	ImageRespository     string            `yaml:"imageRepository"`
	UseHyperKubeImage    bool              `yaml:"useHyperKubeImage"`
	ClusterName          string            `yaml:"clusterName"`
}
type Etcd struct {
	Local map[string]string `yaml:"local"`
}

type Apiserver struct {
	ExtraArgs              map[string]string `yaml:"extraArgs"`
	CertSANs               []string          `yaml:"certSANs"`
	TimeoutForControlPlane string            `yaml:"timeoutForControlPlane"`
}

type KubeProxy struct {
	ApiVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Mode       string `yaml:"mode"`
}

type Kubelet struct {
	ApiVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	MaxPods    string `yaml:"MaxPods"`
}

func DefaultConfig() (*InitCluster, *KubeProxy, *Kubelet) {
	return &InitCluster{
			ApiVersion: "kubeadm.k8s.io/v1beta2",
			Kind:       "ClusterConfiguration",
			Etcd: Etcd{Local: map[string]string{
				"imageRepository": "gcr.azk8s.cn/google_containers",
				"dataDir":         "/data/etcd",
			}},
			Networking: map[string]string{
				"serviceSubnet": config.Configmaps.Kubeconf.ServiceCidr,
				"podSubnet":     config.Configmaps.Kubeconf.PodCidr,
			},
			Dns:                  map[string]string{"type": "CoreDNS", "imageRepository": "coredns"},
			ControlPlaneEndpoint: config.Configmaps.Kubeconf.LoadBalancer,
			ImageRespository:     "gcr.azk8s.cn/google_containers",
			UseHyperKubeImage:    false,
			ClusterName:          "kubernetes",
		}, &KubeProxy{
			ApiVersion: "kubeproxy.config.k8s.io/v1alpha1",
			Kind:       "KubeProxyConfiguration",
			Mode:       config.Configmaps.Kubeconf.ProxyMode,
		}, &Kubelet{
			ApiVersion: "kubelet.config.k8s.io/v1beta1",
			Kind:       "KubeletConfiguration",
			MaxPods:    "500",
		}
}
