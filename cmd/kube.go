package cmd

import (
	"context"
	"fmt"

	"github.com/guobinqiu/appdeployer/docker"
	"github.com/guobinqiu/appdeployer/helpers"
	"github.com/guobinqiu/appdeployer/kube"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type KubeOptions struct {
	Kubeconfig        string
	Namespace         string
	ingressOptions    kube.IngressOptions
	serviceOptions    kube.ServiceOptions
	deploymentOptions kube.DeploymentOptions
}

var dockerOptions docker.DockerOptions
var kubeOptions KubeOptions

func init() {
	// set default values
	viper.SetDefault("docker.dockerconfig", helpers.GetDefaultDockerConfig())
	viper.SetDefault("docker.dockerfile", "./Dockerfile")
	viper.SetDefault("docker.registry", docker.DOCKERHUB)
	viper.SetDefault("docker.tag", "latest")
	viper.SetDefault("kube.kubeconfig", helpers.GetDefaultKubeConfig())
	viper.SetDefault("kube.ingress.tls", true)
	viper.SetDefault("kube.service.port", 8000)
	viper.SetDefault("kube.deployment.replicas", 1)
	viper.SetDefault("kube.deployment.port", 8000)
	viper.SetDefault("kube.deployment.maxsurge", "1")
	viper.SetDefault("kube.deployment.maxunavailable", "0")

	// docker
	kubeCmd.Flags().StringVar(&dockerOptions.Dockerconfig, "docker.dockerconfig", viper.GetString("docker.dockerconfig"), "docker.dockerconfig")
	kubeCmd.Flags().StringVar(&dockerOptions.Dockerfile, "docker.dockerfile", viper.GetString("docker.dockerfile"), "docker.dockerfile")
	kubeCmd.Flags().StringVar(&dockerOptions.Registry, "docker.registry", viper.GetString("docker.registry"), "docker.registry")
	kubeCmd.Flags().StringVar(&dockerOptions.Username, "docker.username", viper.GetString("docker.username"), "docker.username")
	kubeCmd.Flags().StringVar(&dockerOptions.Password, "docker.password", viper.GetString("docker.password"), "docker.password")
	kubeCmd.Flags().StringVar(&dockerOptions.Repository, "docker.repository", viper.GetString("docker.repository"), "docker.repository")
	kubeCmd.Flags().StringVar(&dockerOptions.Tag, "docker.tag", viper.GetString("docker.tag"), "docker.tag")

	//kube
	kubeCmd.Flags().StringVar(&kubeOptions.Kubeconfig, "kube.kubeconfig", viper.GetString("kube.kubeconfig"), "kube.kubeconfig")
	kubeCmd.Flags().StringVar(&kubeOptions.Namespace, "kube.namespace", viper.GetString("kube.namespace"), "kube.namespace")
	kubeCmd.Flags().StringVar(&kubeOptions.ingressOptions.Host, "kube.ingress.host", viper.GetString("kube.ingress.host"), "kube.ingress.host")
	kubeCmd.Flags().BoolVar(&kubeOptions.ingressOptions.TLS, "kube.ingress.tls", viper.GetBool("kube.ingress.tls"), "kube.ingress.tls")
	kubeCmd.Flags().Int32Var(&kubeOptions.serviceOptions.Port, "kube.service.port", viper.GetInt32("kube.service.port"), "kube.service.port")
	kubeCmd.Flags().Int32Var(&kubeOptions.deploymentOptions.Replicas, "kube.deployment.replicas", viper.GetInt32("kube.deployment.replicas"), "kube.deployment.replicas")
	kubeCmd.Flags().Int32Var(&kubeOptions.deploymentOptions.Port, "kube.deployment.port", viper.GetInt32("kube.deployment.port"), "kube.deployment.port")
	kubeCmd.Flags().StringVar(&kubeOptions.deploymentOptions.MaxSurge, "kube.deployment.maxsurge", viper.GetString("kube.deployment.maxsurge"), "kube.deployment.maxsurge")
	kubeCmd.Flags().StringVar(&kubeOptions.deploymentOptions.MaxUnavailable, "kube.deployment.maxunavailable", viper.GetString("kube.deployment.maxunavailable"), "kube.deployment.maxunavailable")
	kubeCmd.Flags().StringVar(&kubeOptions.deploymentOptions.CPULimit, "kube.deployment.cpulimit", viper.GetString("kube.deployment.cpulimit"), "kube.deployment.cpulimit")
	kubeCmd.Flags().StringVar(&kubeOptions.deploymentOptions.MemLimit, "kube.deployment.memlimit", viper.GetString("kube.deployment.memlimit"), "kube.deployment.memlimit")
	kubeCmd.Flags().StringVar(&kubeOptions.deploymentOptions.CPURequest, "kube.deployment.cpurequest", viper.GetString("kube.deployment.cpurequest"), "kube.deployment.cpurequest")
	kubeCmd.Flags().StringVar(&kubeOptions.deploymentOptions.MemRequest, "kube.deployment.memrequest", viper.GetString("kube.deployment.memrequest"), "kube.deployment.memrequest")
}

var kubeCmd = &cobra.Command{
	Use:   "kube",
	Short: "Deploy app to kubernetes cluster",
	Run: func(cmd *cobra.Command, args []string) {
		setDefaultOptions()
		setDockerOptions()
		setKubeOptions()

		gitPull()

		// Create a docker service
		dockerservice, err := docker.NewDockerService()
		if err != nil {
			panic(err)
		}

		//TODO handle timeout or cancel
		ctx := context.TODO()

		// Build an app into a docker image
		if err := dockerservice.BuildImage(ctx, dockerOptions); err != nil {
			panic(err)
		}

		// Push the docker image to docker registry
		if err := dockerservice.PushImage(ctx, dockerOptions); err != nil {
			panic(err)
		}

		dockerservice.Close()

		// Create a kubernetes client by the specified kubeconfig
		config, err := clientcmd.BuildConfigFromFlags("", kubeOptions.Kubeconfig)
		if err != nil {
			panic(err)
		}

		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			panic(err)
		}

		// Update (or create if not exists) kubernetes resource objects
		if err := kube.CreateOrUpdateNamespace(clientset, ctx, kubeOptions.Namespace); err != nil {
			panic(err)
		}

		if err := kube.CreateOrUpdateDockerSecret(clientset, ctx, kube.DockerSecretOptions{
			Name:          defaultOptions.AppName,
			Namespace:     kubeOptions.Namespace,
			DockerOptions: dockerOptions,
		}); err != nil {
			panic(err)
		}

		if err := kube.CreateOrUpdateServiceAccount(clientset, ctx, kube.ServiceAccountOptions{
			Name:      defaultOptions.AppName,
			Namespace: kubeOptions.Namespace,
		}); err != nil {
			panic(err)
		}

		kubeOptions.deploymentOptions.Name = defaultOptions.AppName
		kubeOptions.deploymentOptions.Namespace = kubeOptions.Namespace
		kubeOptions.deploymentOptions.Image = dockerOptions.Image()
		if err := kube.CreateOrUpdateDeployment(clientset, ctx, kubeOptions.deploymentOptions); err != nil {
			panic(err)
		}

		kubeOptions.serviceOptions.Name = defaultOptions.AppName
		kubeOptions.serviceOptions.Namespace = kubeOptions.Namespace
		kubeOptions.serviceOptions.TargetPort = kubeOptions.deploymentOptions.Port
		if err := kube.CreateOrUpdateService(clientset, ctx, kubeOptions.serviceOptions); err != nil {
			panic(err)
		}

		kubeOptions.ingressOptions.Name = defaultOptions.AppName
		kubeOptions.ingressOptions.Namespace = kubeOptions.Namespace
		if err := kube.CreateOrUpdateIngress(clientset, ctx, kubeOptions.ingressOptions); err != nil {
			panic(err)
		}
	},
}

func setDockerOptions() {
	dockerOptions.AppDir = defaultOptions.AppDir

	dockerOptions.Dockerconfig = helpers.ExpandUser(dockerOptions.Dockerconfig)
	exist, err := helpers.IsFileExist(dockerOptions.Dockerconfig)
	if err != nil {
		panic(err)
	}
	if !exist {
		panic("dockerconfig does not exist")
	}

	if helpers.IsBlank(dockerOptions.Repository) && dockerOptions.Registry == docker.DOCKERHUB {
		if helpers.IsBlank(dockerOptions.Username) {
			panic("docker.username is required")
		}
		dockerOptions.Repository = fmt.Sprintf("%s/%s", dockerOptions.Username, defaultOptions.AppName)
	}
}

func setKubeOptions() {
	kubeOptions.Kubeconfig = helpers.ExpandUser(kubeOptions.Kubeconfig)
	exist, err := helpers.IsFileExist(kubeOptions.Kubeconfig)
	if err != nil {
		panic(err)
	}
	if !exist {
		panic("kubeconfig does not exist")
	}

	if helpers.IsBlank(kubeOptions.Namespace) {
		kubeOptions.Namespace = defaultOptions.AppName
	}

	if helpers.IsBlank(kubeOptions.ingressOptions.Host) {
		kubeOptions.ingressOptions.Host = fmt.Sprintf("%s.com", defaultOptions.AppName)
	}
}
