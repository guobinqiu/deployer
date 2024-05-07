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
	hpaOptions        kube.HPAOptions
	pvcOptions        kube.PVCOptions
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

	viper.SetDefault("kube.ingress.tls", false)
	viper.SetDefault("kube.ingress.selfsigned", false)
	viper.SetDefault("kube.ingress.selfsignedyears", 1)

	viper.SetDefault("kube.service.port", 8000)

	viper.SetDefault("kube.deployment.replicas", 1)
	viper.SetDefault("kube.deployment.port", 8000)

	viper.SetDefault("kube.deployment.rollingupdate.maxsurge", "1")
	viper.SetDefault("kube.deployment.rollingupdate.maxunavailable", "0")

	viper.SetDefault("kube.deployment.livenessprobe.enabled", false)
	viper.SetDefault("kube.deployment.livenessprobe.type", kube.ProbeTypeHTTPGet)
	viper.SetDefault("kube.deployment.livenessprobe.path", "/")
	viper.SetDefault("kube.deployment.livenessprobe.schema", "http")
	viper.SetDefault("kube.deployment.livenessprobe.initialdelayseconds", 0)
	viper.SetDefault("kube.deployment.livenessprobe.timeoutseconds", 1)
	viper.SetDefault("kube.deployment.livenessprobe.periodseconds", 10)
	viper.SetDefault("kube.deployment.livenessprobe.successthreshold", 1)
	viper.SetDefault("kube.deployment.livenessprobe.failurethreshold", 3)

	viper.SetDefault("kube.deployment.readinessprobe.enabled", false)
	viper.SetDefault("kube.deployment.readinessprobe.type", kube.ProbeTypeHTTPGet)
	viper.SetDefault("kube.deployment.readinessprobe.path", "/")
	viper.SetDefault("kube.deployment.readinessprobe.schema", "http")
	viper.SetDefault("kube.deployment.readinessprobe.initialdelayseconds", 0)
	viper.SetDefault("kube.deployment.readinessprobe.timeoutseconds", 1)
	viper.SetDefault("kube.deployment.readinessprobe.periodseconds", 10)
	viper.SetDefault("kube.deployment.readinessprobe.successthreshold", 1)
	viper.SetDefault("kube.deployment.readinessprobe.failurethreshold", 3)

	viper.SetDefault("kube.deployment.volumemount.enabled", false)
	viper.SetDefault("kube.deployment.volumemount.mountpath", "/app/data")

	viper.SetDefault("kube.hpa.enabled", false)
	viper.SetDefault("kube.hpa.minreplicas", 1)
	viper.SetDefault("kube.hpa.maxreplicas", 10)
	viper.SetDefault("kube.hpa.cpurate", 50)

	viper.SetDefault("kube.pvc.accessmode", "readwriteonce")
	viper.SetDefault("kube.pvc.storageclassname", "openebs-hostpath")
	viper.SetDefault("kube.pvc.storagesize", "1G")

	// docker
	kubeCmd.Flags().StringVar(&dockerOptions.Dockerconfig, "docker.dockerconfig", viper.GetString("docker.dockerconfig"), "Path to docker configuration. Defaults to ~/.docker/config.json")
	kubeCmd.Flags().StringVar(&dockerOptions.Dockerfile, "docker.dockerfile", viper.GetString("docker.dockerfile"), "Path to Dockerfile for building image. Defaults to appdir/Dockerfile")
	kubeCmd.Flags().StringVar(&dockerOptions.Registry, "docker.registry", viper.GetString("docker.registry"), "URL for docker registry. Defaults to https://index.docker.io/v1/")
	kubeCmd.Flags().StringVar(&dockerOptions.Username, "docker.username", viper.GetString("docker.username"), "Username for docker registry")
	kubeCmd.Flags().StringVar(&dockerOptions.Password, "docker.password", viper.GetString("docker.password"), "Password for docker registry")
	kubeCmd.Flags().StringVar(&dockerOptions.Repository, "docker.repository", viper.GetString("docker.repository"), "Repository for docker registry")
	kubeCmd.Flags().StringVar(&dockerOptions.Tag, "docker.tag", viper.GetString("docker.tag"), "Tag for docker registry. Defaults to latest")

	//kube
	kubeCmd.Flags().StringVar(&kubeOptions.Kubeconfig, "kube.kubeconfig", viper.GetString("kube.kubeconfig"), "Path to kubernetes configuration. Defaults to ~/.kube/config")
	kubeCmd.Flags().StringVar(&kubeOptions.Namespace, "kube.namespace", viper.GetString("kube.namespace"), "Namespace for app resources. Defaults to appname")

	kubeCmd.Flags().StringVar(&kubeOptions.ingressOptions.Host, "kube.ingress.host", viper.GetString("kube.ingress.host"), "Host for app ingress. Defaults to appName.com")
	kubeCmd.Flags().BoolVar(&kubeOptions.ingressOptions.TLS, "kube.ingress.tls", viper.GetBool("kube.ingress.tls"), "Enable or disable TLS for app host. Defaults to false")
	kubeCmd.Flags().BoolVar(&kubeOptions.ingressOptions.SelfSigned, "kube.ingress.selfsigned", viper.GetBool("kube.ingress.selfsigned"), "Enable or disable self-signed certificate. Defaults to false")
	kubeCmd.Flags().IntVar(&kubeOptions.ingressOptions.SelfSignedYears, "kube.ingress.selfsignedyears", viper.GetInt("kube.ingress.selfsignedyears"), "Validity of self-signed certificate. Defaults to 1 year")
	kubeCmd.Flags().StringVar(&kubeOptions.ingressOptions.CrtPath, "kube.ingress.crtpath", viper.GetString("kube.ingress.crtpath"), "Path to .crt file (PEM format) for non self-signed certificate")
	kubeCmd.Flags().StringVar(&kubeOptions.ingressOptions.KeyPath, "kube.ingress.keypath", viper.GetString("kube.ingress.keypath"), "Path to .key file (PEM format) for non self-signed certificate")

	kubeCmd.Flags().Int32Var(&kubeOptions.serviceOptions.Port, "kube.service.port", viper.GetInt32("kube.service.port"), "Port for app service. Defaults to 8000")

	kubeCmd.Flags().Int32Var(&kubeOptions.deploymentOptions.Replicas, "kube.deployment.replicas", viper.GetInt32("kube.deployment.replicas"), "Number of app pods. Defaults to 1")
	kubeCmd.Flags().Int32Var(&kubeOptions.deploymentOptions.Port, "kube.deployment.port", viper.GetInt32("kube.deployment.port"), "Container port for each app pod. Defaults to 8000, as same as service port")
	kubeCmd.Flags().StringVar(&kubeOptions.deploymentOptions.RollingUpdate.MaxSurge, "kube.deployment.rollingupdate.maxsurge", viper.GetString("kube.deployment.rollingupdate.maxsurge"), "MaxSurge for rolling update app pods. Defaults to 1")
	kubeCmd.Flags().StringVar(&kubeOptions.deploymentOptions.RollingUpdate.MaxUnavailable, "kube.deployment.rollingupdate.maxunavailable", viper.GetString("kube.deployment.rollingupdate.maxunavailable"), "MaxUnavailable for rolling update app pods. Defaults to 0")
	kubeCmd.Flags().StringVar(&kubeOptions.deploymentOptions.Quota.CPULimit, "kube.deployment.quota.cpulimit", viper.GetString("kube.deployment.quota.cpulimit"), "CPU limit for each app container (one pod one container)")
	kubeCmd.Flags().StringVar(&kubeOptions.deploymentOptions.Quota.MemLimit, "kube.deployment.quota.memlimit", viper.GetString("kube.deployment.quota.memlimit"), "Memory limit for each app container (one pod one container)")
	kubeCmd.Flags().StringVar(&kubeOptions.deploymentOptions.Quota.CPURequest, "kube.deployment.quota.cpurequest", viper.GetString("kube.deployment.quota.cpurequest"), "CPU request for each app container (one pod one container)")
	kubeCmd.Flags().StringVar(&kubeOptions.deploymentOptions.Quota.MemRequest, "kube.deployment.quota.memrequest", viper.GetString("kube.deployment.quota.memrequest"), "Memory request for each app container (one pod one container)")

	kubeCmd.Flags().StringSliceVarP(&kubeOptions.deploymentOptions.EnvVars, "env", "e", nil, "Set environment variables in the form of key=value")

	kubeCmd.Flags().BoolVar(&kubeOptions.deploymentOptions.LivenessProbe.Enabled, "kube.deployment.livenessprobe.enabled", viper.GetBool("kube.deployment.livenessprobe.enabled"), "Enable or disable liveness probe for each app container (one pod one container). Defaults to false")
	kubeCmd.Flags().StringVar(&kubeOptions.deploymentOptions.LivenessProbe.Type, "kube.deployment.livenessprobe.type", viper.GetString("kube.deployment.livenessprobe.type"), "Type of liveness probe for each app container (one pod one container). Such as HTTPGet, TCPSocket and Exec. Defaults to HTTPGet")
	kubeCmd.Flags().StringVar(&kubeOptions.deploymentOptions.LivenessProbe.Path, "kube.deployment.livenessprobe.path", viper.GetString("kube.deployment.livenessprobe.path"), "Path of liveness probe for each app container (one pod one container). Correspond to HTTPGet type. Defaults to /")
	kubeCmd.Flags().StringVar(&kubeOptions.deploymentOptions.LivenessProbe.Schema, "kube.deployment.livenessprobe.schema", viper.GetString("kube.deployment.livenessprobe.schema"), "Schema of liveness probe for each app container (one pod one container). Correspond to HTTPGet type. Such as HTTP and HTTPS. Defaults to HTTP")
	kubeCmd.Flags().StringVar(&kubeOptions.deploymentOptions.LivenessProbe.Command, "kube.deployment.livenessprobe.command", viper.GetString("kube.deployment.livenessprobe.command"), "Command of liveness probe for each app container (one pod one container). Correspond to Exec type")
	kubeCmd.Flags().Int32Var(&kubeOptions.deploymentOptions.LivenessProbe.InitialDelaySeconds, "kube.deployment.livenessprobe.initialdelayseconds", viper.GetInt32("kube.deployment.livenessprobe.initialdelayseconds"), "Initial delay seconds of liveness probe for each app container (one pod one container). Defaults to 0")
	kubeCmd.Flags().Int32Var(&kubeOptions.deploymentOptions.LivenessProbe.TimeoutSeconds, "kube.deployment.livenessprobe.timeoutseconds", viper.GetInt32("kube.deployment.livenessprobe.timeoutseconds"), "Timeout seconds of liveness probe for each app container (one pod one container). Defaults to 1")
	kubeCmd.Flags().Int32Var(&kubeOptions.deploymentOptions.LivenessProbe.PeriodSeconds, "kube.deployment.livenessprobe.periodseconds", viper.GetInt32("kube.deployment.livenessprobe.periodseconds"), "Period seconds of liveness probe for each app container (one pod one container). Defaults to 10")
	kubeCmd.Flags().Int32Var(&kubeOptions.deploymentOptions.LivenessProbe.SuccessThreshold, "kube.deployment.livenessprobe.successthreshold", viper.GetInt32("kube.deployment.livenessprobe.successthreshold"), "Success threshold of liveness probe for each app container (one pod one container). Defaults to 1")
	kubeCmd.Flags().Int32Var(&kubeOptions.deploymentOptions.LivenessProbe.FailureThreshold, "kube.deployment.livenessprobe.failurethreshold", viper.GetInt32("kube.deployment.livenessprobe.failurethreshold"), "Failure threshold of liveness probe for each app container (one pod one container). Defaults to 3")

	kubeCmd.Flags().BoolVar(&kubeOptions.deploymentOptions.ReadinessProbe.Enabled, "kube.deployment.readinessprobe.enabled", viper.GetBool("kube.deployment.readinessprobe.enabled"), "Enable or disable readiness probe for each app container (one pod one container)")
	kubeCmd.Flags().StringVar(&kubeOptions.deploymentOptions.ReadinessProbe.Type, "kube.deployment.readinessprobe.type", viper.GetString("kube.deployment.readinessprobe.type"), "Type of readiness probe for each app container (one pod one container). Such as HTTPGet, TCPSocket and Exec. Defaults to HTTPGet")
	kubeCmd.Flags().StringVar(&kubeOptions.deploymentOptions.ReadinessProbe.Path, "kube.deployment.readinessprobe.path", viper.GetString("kube.deployment.readinessprobe.path"), "Path of readiness probe for each app container (one pod one container). Correspond to HTTPGet type. Defaults to /")
	kubeCmd.Flags().StringVar(&kubeOptions.deploymentOptions.ReadinessProbe.Schema, "kube.deployment.readinessprobe.schema", viper.GetString("kube.deployment.readinessprobe.schema"), "Schema of readiness probe for each app container (one pod one container). Correspond to HTTPGet type. Such as HTTP and HTTPS. Defaults to HTTP")
	kubeCmd.Flags().StringVar(&kubeOptions.deploymentOptions.ReadinessProbe.Command, "kube.deployment.readinessprobe.command", viper.GetString("kube.deployment.readinessprobe.command"), "Command of readiness probe for each app container (one pod one container). Correspond to Exec type")
	kubeCmd.Flags().Int32Var(&kubeOptions.deploymentOptions.ReadinessProbe.InitialDelaySeconds, "kube.deployment.readinessprobe.initialdelayseconds", viper.GetInt32("kube.deployment.readinessprobe.initialdelayseconds"), "Initial delay seconds of readiness probe for each app container (one pod one container). Defaults to 0")
	kubeCmd.Flags().Int32Var(&kubeOptions.deploymentOptions.ReadinessProbe.TimeoutSeconds, "kube.deployment.readinessprobe.timeoutseconds", viper.GetInt32("kube.deployment.readinessprobe.timeoutseconds"), "Timeout seconds of readiness probe for each app container (one pod one container). Defaults to 1")
	kubeCmd.Flags().Int32Var(&kubeOptions.deploymentOptions.ReadinessProbe.PeriodSeconds, "kube.deployment.readinessprobe.periodseconds", viper.GetInt32("kube.deployment.readinessprobe.periodseconds"), "Period seconds of readiness probe for each app container (one pod one container). Defaults to 10")
	kubeCmd.Flags().Int32Var(&kubeOptions.deploymentOptions.ReadinessProbe.SuccessThreshold, "kube.deployment.readinessprobe.successthreshold", viper.GetInt32("kube.deployment.readinessprobe.successthreshold"), "Success threshold of readiness probe for each app container (one pod one container). Defaults to 1")
	kubeCmd.Flags().Int32Var(&kubeOptions.deploymentOptions.ReadinessProbe.FailureThreshold, "kube.deployment.readinessprobe.failurethreshold", viper.GetInt32("kube.deployment.readinessprobe.failurethreshold"), "Failure threshold of readiness probe for each app container (one pod one container). Defaults to 3")

	kubeCmd.Flags().BoolVar(&kubeOptions.deploymentOptions.VolumeMount.Enabled, "kube.deployment.volumemount.enabled", viper.GetBool("kube.deployment.volumemount.enabled"), "Enable or disable volume mount for each app pod. Defaults to false")
	kubeCmd.Flags().StringVar(&kubeOptions.deploymentOptions.VolumeMount.MountPath, "kube.deployment.volumemount.mountpath", viper.GetString("kube.deployment.volumemount.mountpath"), "Path of volume mount for each app pod. Defaults to /app/data")

	kubeCmd.Flags().BoolVar(&kubeOptions.hpaOptions.Enabled, "kube.hpa.enabled", viper.GetBool("kube.hpa.enabled"), "Enable or disable HPA (Horizontal Pod Autoscaler) for app pods. Defaults to false")
	kubeCmd.Flags().Int32Var(&kubeOptions.hpaOptions.MinReplicas, "kube.hpa.minreplicas", viper.GetInt32("kube.hpa.minreplicas"), "Number of minimum pods for HPA (Horizontal Pod Autoscaler). Defaults to 1")
	kubeCmd.Flags().Int32Var(&kubeOptions.hpaOptions.MaxReplicas, "kube.hpa.maxreplicas", viper.GetInt32("kube.hpa.maxreplicas"), "Number of maximum pods for HPA (Horizontal Pod Autoscaler). Defaults to 10")
	kubeCmd.Flags().Int32Var(&kubeOptions.hpaOptions.CPURate, "kube.hpa.cpurate", viper.GetInt32("kube.hpa.cpurate"), "Average CPU utilization for HPA (Horizontal Pod Autoscaler). Defaults to 50")

	kubeCmd.Flags().StringVar(&kubeOptions.pvcOptions.AccessMode, "kube.pvc.accessmode", viper.GetString("kube.pvc.accessmode"), "Access mode of persistent storage for pod volumn mount. Such as ReadWriteOnce, ReadOnlyMany and ReadWriteMany. Defaults to ReadWriteOnce")
	kubeCmd.Flags().StringVar(&kubeOptions.pvcOptions.StorageClassName, "kube.pvc.storageclassname", viper.GetString("kube.pvc.storageclassname"), "Classname of persistent storage for pod volumn mount. Defaults to openebs-hostpath")
	kubeCmd.Flags().StringVar(&kubeOptions.pvcOptions.StorageSize, "kube.pvc.storagesize", viper.GetString("kube.pvc.storagesize"), "Size of persistent storage for pod volumn mount. Defaults to 1G")
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

		// Update or create kubernetes resource objects
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

		if kubeOptions.deploymentOptions.VolumeMount.Enabled {
			kubeOptions.pvcOptions.Name = defaultOptions.AppName
			kubeOptions.pvcOptions.Namespace = kubeOptions.Namespace
			if err := kube.CreateOrUpdatePVC(clientset, ctx, kubeOptions.pvcOptions); err != nil {
				panic(err)
			}
		} else {
			kubeOptions.deploymentOptions.Name = defaultOptions.AppName
			kubeOptions.deploymentOptions.Namespace = kubeOptions.Namespace
			kubeOptions.deploymentOptions.Image = dockerOptions.Image()
			if err := kube.DeleteDeployment(clientset, ctx, kubeOptions.deploymentOptions); err != nil {
				panic(err)
			}

			kubeOptions.hpaOptions.Name = defaultOptions.AppName
			kubeOptions.hpaOptions.Namespace = kubeOptions.Namespace
			if err := kube.DeletePVC(clientset, ctx, kubeOptions.hpaOptions); err != nil {
				panic(err)
			}
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

		if kubeOptions.hpaOptions.Enabled {
			kubeOptions.hpaOptions.Name = defaultOptions.AppName
			kubeOptions.hpaOptions.Namespace = kubeOptions.Namespace
			if err := kube.CreateOrUpdateHPA(clientset, ctx, kubeOptions.hpaOptions); err != nil {
				panic(err)
			}
		} else {
			kubeOptions.hpaOptions.Name = defaultOptions.AppName
			kubeOptions.hpaOptions.Namespace = kubeOptions.Namespace
			if err := kube.DeleteHPA(clientset, ctx, kubeOptions.hpaOptions); err != nil {
				panic(err)
			}
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

	if kubeOptions.ingressOptions.TLS && !kubeOptions.ingressOptions.SelfSigned {
		if helpers.IsBlank(kubeOptions.ingressOptions.CrtPath) {
			panic("crt path does not exist")
		}
		if helpers.IsBlank(kubeOptions.ingressOptions.KeyPath) {
			panic("key path does not exist")
		}
	}
}
