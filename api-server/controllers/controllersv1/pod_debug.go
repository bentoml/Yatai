package controllersv1

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"strconv"
	"sync"
	"time"

	authorizationv1 "k8s.io/api/authorization/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	coreclient "k8s.io/client-go/kubernetes/typed/core/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/tools/watch"
	"k8s.io/client-go/transport/spdy"
	"k8s.io/kubernetes/pkg/client/conditions"
	"k8s.io/kubernetes/pkg/util/interrupt"

	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/api-server/transformers/transformersv1"
	"github.com/bentoml/yatai/common/consts"
)

const (
	defaultImage         = "docker.io/nicolaka/netshoot:latest"
	defaultAgentPort     = 10027
	defaultDaemonSetName = "debug-agent"
	defaultDaemonSetNs   = "default"

	defaultAgentImage             = "aylei/debug-agent:latest"
	defaultAgentPodNamePrefix     = "debug-agent-pod"
	defaultAgentPodNamespace      = "default"
	defaultAgentPodCpuRequests    = ""
	defaultAgentPodCpuLimits      = ""
	defaultAgentPodMemoryRequests = ""
	defaultAgentPodMemoryLimits   = ""

	// nolint: gosec
	defaultRegistrySecretName      = "kubectl-debug-registry-secret"
	defaultRegistrySecretNamespace = "default"
	defaultRegistrySkipTLSVerify   = false

	defaultLxcfsEnable = true
	defaultVerbosity   = 0
)

// DebugOptions specify how to run debug container in a running pod
type DebugOptions struct {

	// Pod select options
	Namespace string
	PodName   string

	// Debug options
	Image                   string
	RegistrySecretName      string
	RegistrySecretNamespace string
	RegistrySkipTLSVerify   bool

	ContainerName       string
	Command             []string
	AgentPort           int
	AppName             string
	ConfigLocation      string
	Fork                bool
	ForkPodRetainLabels []string
	// used for agentless mode
	AgentLess  bool
	AgentImage string
	// agentPodName = agentPodNamePrefix + nodeName
	AgentPodName      string
	AgentPodNamespace string
	AgentPodNode      string
	AgentPodResource  agentPodResources
	// enable lxcfs
	IsLxcfsEnabled bool

	Flags      *genericclioptions.ConfigFlags
	CoreClient coreclient.CoreV1Interface
	KubeCli    *kubernetes.Clientset
	Args       []string
	Config     *restclient.Config

	// use for port-forward
	RESTClient    *restclient.RESTClient
	PortForwarder portForwarder
	Ports         []string
	StopChannel   chan struct{}
	ReadyChannel  chan struct{}

	PortForward         bool
	DebugAgentDaemonSet string
	DebugAgentNamespace string

	genericclioptions.IOStreams

	wait sync.WaitGroup

	Verbosity int
	Logger    *log.Logger

	Terminal *WebTerminal
}

type agentPodResources struct {
	CpuRequests    string
	CpuLimits      string
	MemoryRequests string
	MemoryLimits   string
}

// NewDebugOptions new debug options
func NewDebugOptions(terminal *WebTerminal, clientset *kubernetes.Clientset, restConfig *restclient.Config, agentLess, fork bool, image string) *DebugOptions {
	streams := genericclioptions.IOStreams{
		In:     terminal,
		Out:    terminal,
		ErrOut: terminal,
	}

	o := &DebugOptions{
		Terminal:  terminal,
		IOStreams: streams,
		PortForwarder: &defaultPortForwarder{
			IOStreams: streams,
		},
		Logger:                  log.New(streams.Out, "kubectl-debug ", log.LstdFlags|log.Lshortfile),
		Config:                  restConfig,
		KubeCli:                 clientset,
		CoreClient:              clientset.CoreV1(),
		StopChannel:             make(chan struct{}, 1),
		ReadyChannel:            make(chan struct{}),
		Namespace:               terminal.namespace,
		PodName:                 terminal.podName,
		ContainerName:           terminal.containerName,
		AgentLess:               agentLess,
		Fork:                    fork,
		ForkPodRetainLabels:     []string{consts.KubeLabelYataiDeployment},
		Command:                 []string{"bash", "-l"},
		Image:                   image,
		RegistrySecretName:      defaultRegistrySecretName,
		RegistrySecretNamespace: defaultRegistrySecretNamespace,
		RegistrySkipTLSVerify:   defaultRegistrySkipTLSVerify,
		AgentPort:               defaultAgentPort,
		Verbosity:               defaultVerbosity,
		DebugAgentNamespace:     defaultDaemonSetNs,
		DebugAgentDaemonSet:     defaultDaemonSetName,
		AgentPodName:            defaultAgentPodNamePrefix,
		AgentImage:              defaultAgentImage,
		AgentPodNamespace:       defaultAgentPodNamespace,
	}
	o.AgentPodResource.CpuRequests = defaultAgentPodCpuRequests
	o.AgentPodResource.MemoryRequests = defaultAgentPodMemoryRequests
	o.AgentPodResource.CpuLimits = defaultAgentPodCpuLimits
	o.AgentPodResource.MemoryLimits = defaultAgentPodMemoryLimits
	o.IsLxcfsEnabled = defaultLxcfsEnable
	o.Ports = []string{strconv.Itoa(o.AgentPort)}
	if o.Image == "" {
		o.Image = defaultImage
	}
	return o
}

// Validate validate
func (o *DebugOptions) Validate() error {
	if len(o.PodName) == 0 {
		return fmt.Errorf("pod name must be specified")
	}
	if len(o.Command) == 0 {
		return fmt.Errorf("you must specify at least one command for the container")
	}
	return nil
}

// TODO: refactor Run() spaghetti code
// Run run
func (o *DebugOptions) Run(ctx context.Context) error {
	pod, err := o.CoreClient.Pods(o.Namespace).Get(ctx, o.PodName, v1.GetOptions{})
	if err != nil {
		return err
	}

	containerName := o.ContainerName
	if len(containerName) == 0 {
		if len(pod.Spec.Containers) > 1 {
			usageString := fmt.Sprintf("Defaulting container name to %s.", pod.Spec.Containers[0].Name)
			fmt.Fprintf(o.ErrOut, "%s\n\r", usageString)
		}
		containerName = pod.Spec.Containers[0].Name
	}
	err = o.auth(ctx, pod)
	if err != nil {
		return err
	}
	// Launch debug launching pod in agentless mode.
	var agentPod *corev1.Pod
	if o.AgentLess {
		o.AgentPodNode = pod.Spec.NodeName
		o.AgentPodName = fmt.Sprintf("%s-%s", o.AgentPodName, uuid.NewUUID())
		agentPod = o.getAgentPod()
		agentPod, err = o.launchPod(ctx, agentPod)
		if err != nil {
			fmt.Fprintf(o.Out, "the agentPod is not running, you should check the reason and delete the failed agentPod and retry.\n")
			return err
		}
	}

	// in fork mode, we launch an new pod as a copy of target pod
	// and hack the entry point of the target container with sleep command
	// which keeps the container running.
	if o.Fork {
		// build the fork pod labels
		podLabels := o.buildForkPodLabels(pod)
		// copy pod and run
		pod = copyAndStripPod(pod, containerName, podLabels)
		pod_ := services.KubePodService.MapKubePodsToKubePodWithStatuses(ctx, []corev1.Pod{*pod}, nil)[0]
		podView, err := transformersv1.ToKubePodSchema(ctx, pod_)
		if err != nil {
			return err
		}
		_ = o.Terminal.conn.WriteJSON(map[string]interface{}{
			"is_mcd_msg": true,
			"pod":        podView,
		})
		pod, err = o.launchPod(ctx, pod)
		if err != nil {
			fmt.Fprintf(o.Out, "the ForkedPod is not running, you should check the reason and delete the failed ForkedPod and retry\n")
			o.deleteAgent(ctx, agentPod)
			return err
		}
	}

	if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
		o.deleteAgent(ctx, agentPod)
		return fmt.Errorf("cannot debug in a completed pod; current phase is %s", pod.Status.Phase)
	}

	containerID, err := o.getContainerIDByName(pod, containerName)
	if err != nil {
		o.deleteAgent(ctx, agentPod)
		return err
	}

	if o.PortForward {
		var agent *corev1.Pod
		if !o.AgentLess {
			// Agent is running
			daemonSet, err := o.KubeCli.AppsV1().DaemonSets(o.DebugAgentNamespace).Get(ctx, o.DebugAgentDaemonSet, v1.GetOptions{})
			if err != nil {
				return err
			}
			labelSet := labels.Set(daemonSet.Spec.Selector.MatchLabels)
			agents, err := o.CoreClient.Pods(o.DebugAgentNamespace).List(ctx, v1.ListOptions{
				LabelSelector: labelSet.String(),
			})
			if err != nil {
				return err
			}
			for i := range agents.Items {
				if agents.Items[i].Spec.NodeName == pod.Spec.NodeName {
					agent = &agents.Items[i]
					break
				}
			}
		} else {
			agent = agentPod
		}

		if agent == nil {
			return fmt.Errorf("there is no agent pod in the same node with your speficy pod %s", o.PodName)
		}
		if o.Verbosity > 0 {
			fmt.Fprintf(o.Out, "pod %s PodIP %s, agentPodIP %s\n", o.PodName, pod.Status.PodIP, agent.Status.HostIP)
		}
		err = o.runPortForward(agent)
		if err != nil {
			o.deleteAgent(ctx, agentPod)
			return err
		}
		// client can't access the node ip in the k8s cluster sometimes,
		// than we use forward ports to connect the specified pod and that will listen
		// on specified ports in localhost, the ports can not access until receive the
		// ready signal
		if o.Verbosity > 0 {
			fmt.Fprintln(o.Out, "wait for forward port to debug agent ready...")
		}
		<-o.ReadyChannel
	}

	fn := func() error {
		// TODO: refactor as kubernetes api style, reuse rbac mechanism of kubernetes
		var targetHost string
		if o.PortForward {
			targetHost = "localhost"
		} else {
			targetHost = pod.Status.HostIP
		}
		uri, err := url.Parse(fmt.Sprintf("http://%s:%d", targetHost, o.AgentPort))
		if err != nil {
			return err
		}
		uri.Path = "/api/v1/debug"
		params := url.Values{}
		params.Add("image", o.Image)
		params.Add("container", containerID)
		// FIXME: if verbosity = 0 kubectl-debug agent cannot pull image
		params.Add("verbosity", fmt.Sprintf("%v", 1))
		hstNm, _ := os.Hostname()
		params.Add("hostname", hstNm)
		usr, _ := user.Current()
		params.Add("username", usr.Username)
		if o.IsLxcfsEnabled {
			params.Add("lxcfsEnabled", "true")
		} else {
			params.Add("lxcfsEnabled", "false")
		}
		if o.RegistrySkipTLSVerify {
			params.Add("registrySkipTLS", "true")
		} else {
			params.Add("registrySkipTLS", "false")
		}
		var authStr string
		registrySecret, err := o.CoreClient.Secrets(o.RegistrySecretNamespace).Get(ctx, o.RegistrySecretName, v1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				authStr = ""
			} else {
				return err
			}
		} else {
			authStr = string(registrySecret.Data["authStr"])
		}
		params.Add("authStr", authStr)
		if len(o.Command) > 0 {
			commandBytes, err := json.Marshal(o.Command)
			if err != nil {
				return err
			}
			params.Add("command", string(commandBytes))
		}
		uri.RawQuery = params.Encode()
		return o.remoteExecute("POST", uri, o.Config, o.In, o.Out, o.ErrOut, true, o.Terminal)
	}

	// ensure forked pod is deleted on cancelation
	withCleanUp := func() error {
		return interrupt.Chain(nil, func() {
			if o.Fork {
				fmt.Fprintf(o.Out, "Start deleting forked pod %s \n\r", pod.Name)
				err := o.CoreClient.Pods(pod.Namespace).Delete(ctx, pod.Name, *v1.NewDeleteOptions(0))
				if err != nil {
					// we may leak pod here, but we have nothing to do except noticing the user
					fmt.Fprintf(o.ErrOut, "failed to delete forked pod[Name:%s, Namespace:%s], consider manual deletion.\n\r", pod.Name, pod.Namespace)
				}
			}

			if o.PortForward {
				// close the port-forward
				if o.StopChannel != nil {
					close(o.StopChannel)
				}
			}
			// delete agent pod
			if o.AgentLess && agentPod != nil {
				fmt.Fprintf(o.Out, "Start deleting agent pod %s \n\r", pod.Name)
				o.deleteAgent(ctx, agentPod)
			}
		}).Run(fn)
	}

	if err := o.Terminal.Safe(withCleanUp); err != nil {
		fmt.Fprintf(o.Out, "error execute remote, %v\n", err)
		return err
	}
	o.wait.Wait()
	return nil
}

func (o *DebugOptions) getContainerIDByName(pod *corev1.Pod, containerName string) (string, error) {
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.Name != containerName {
			continue
		}
		// #52 if a pod is running but not ready(because of readiness probe), we can connect
		if containerStatus.State.Running == nil {
			return "", fmt.Errorf("container [%s] not running", containerName)
		}
		if o.Verbosity > 0 {
			o.Logger.Printf("Getting id from containerStatus %+v\r\n", containerStatus)
		}
		return containerStatus.ContainerID, nil
	}

	// #14 otherwise we should search for running init containers
	for _, initContainerStatus := range pod.Status.InitContainerStatuses {
		if initContainerStatus.Name != containerName {
			continue
		}
		if initContainerStatus.State.Running == nil {
			return "", fmt.Errorf("init container [%s] is not running", containerName)
		}
		if o.Verbosity > 0 {
			o.Logger.Printf("Getting id from initContainerStatus %+v\r\n", initContainerStatus)
		}
		return initContainerStatus.ContainerID, nil
	}

	return "", fmt.Errorf("cannot find specified container %s", containerName)
}

func (o *DebugOptions) remoteExecute(
	method string,
	url *url.URL,
	config *restclient.Config,
	stdin io.Reader,
	stdout, stderr io.Writer,
	tty bool,
	terminalSizeQueue remotecommand.TerminalSizeQueue) error {
	if o.Verbosity > 0 {
		o.Logger.Printf("Creating SPDY executor %+v %+v %+v\r\n", config, method, url)
	}
	exec, err := remotecommand.NewSPDYExecutor(config, method, url)
	if err != nil {
		o.Logger.Printf("Error creating SPDY executor.\r\n")
		return err
	}
	if o.Verbosity > 0 {
		o.Logger.Printf("Creating exec Stream\r\n")
	}
	return exec.Stream(remotecommand.StreamOptions{
		Stdin:             stdin,
		Stdout:            stdout,
		Stderr:            stderr,
		Tty:               tty,
		TerminalSizeQueue: terminalSizeQueue,
	})
}

func (o *DebugOptions) buildForkPodLabels(pod *corev1.Pod) map[string]string {
	podLabels := map[string]string{}
	for _, label := range o.ForkPodRetainLabels {
		for k, v := range pod.ObjectMeta.Labels {
			if label == k {
				podLabels[k] = v
			}
		}
	}
	return podLabels
}

// copyAndStripPod copy the given pod template, strip the probes and labels,
// and replace the entry point
func copyAndStripPod(pod *corev1.Pod, targetContainer string, podLabels map[string]string) *corev1.Pod {
	copied := &corev1.Pod{
		ObjectMeta: *pod.ObjectMeta.DeepCopy(),
		Spec:       *pod.Spec.DeepCopy(),
	}
	copied.Name = fmt.Sprintf("%s-%s-debug", pod.Name, uuid.NewUUID())
	copied.Labels = podLabels
	copied.Spec.RestartPolicy = corev1.RestartPolicyNever
	for i, c := range copied.Spec.Containers {
		copied.Spec.Containers[i].LivenessProbe = nil
		copied.Spec.Containers[i].ReadinessProbe = nil
		if c.Name == targetContainer {
			// Hack, infinite sleep command to keep the container running
			copied.Spec.Containers[i].Command = []string{"sh", "-c", "--"}
			copied.Spec.Containers[i].Args = []string{"while true; do sleep 30; done;"}
		}
	}
	copied.ResourceVersion = ""
	copied.UID = ""
	copied.SelfLink = ""
	copied.CreationTimestamp = v1.Time{}
	copied.OwnerReferences = []v1.OwnerReference{}

	return copied
}

// launchPod launch given pod until it's running
func (o *DebugOptions) launchPod(ctx context.Context, pod *corev1.Pod) (*corev1.Pod, error) {
	pod, err := o.CoreClient.Pods(pod.Namespace).Create(ctx, pod, v1.CreateOptions{})
	if err != nil {
		return pod, err
	}

	watcher, err := o.CoreClient.Pods(pod.Namespace).Watch(ctx, v1.SingleObject(pod.ObjectMeta))
	if err != nil {
		return nil, err
	}
	// FIXME: hard code -> config
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	fmt.Fprintf(o.Out, "Waiting for pod %s to run...\n", pod.Name)
	event, err := watch.UntilWithoutRetry(ctx, watcher, conditions.PodRunning)
	if err != nil {
		fmt.Fprintf(o.ErrOut, "Error occurred while waiting for pod to run:  %v\n", err)
		return nil, err
	}
	pod = event.Object.(*corev1.Pod)
	return pod, nil
}

// getAgentPod construnct agentPod from agent pod template
func (o *DebugOptions) getAgentPod() *corev1.Pod {
	prop := corev1.MountPropagationBidirectional
	directoryCreate := corev1.HostPathDirectoryOrCreate
	privileged := true
	agentPod := &corev1.Pod{
		TypeMeta: v1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      o.AgentPodName,
			Namespace: o.AgentPodNamespace,
		},
		Spec: corev1.PodSpec{
			HostPID:  true,
			NodeName: o.AgentPodNode,
			Containers: []corev1.Container{
				{
					Name:            "debug-agent",
					Image:           o.AgentImage,
					ImagePullPolicy: corev1.PullAlways,
					LivenessProbe: &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							HTTPGet: &corev1.HTTPGetAction{
								Path: "/healthz",
								Port: intstr.FromInt(10027),
							},
						},
						InitialDelaySeconds: 10,
						PeriodSeconds:       10,
						SuccessThreshold:    1,
						TimeoutSeconds:      1,
						FailureThreshold:    3,
					},
					SecurityContext: &corev1.SecurityContext{
						Privileged: &privileged,
					},
					Resources: o.buildAgentResourceRequirements(),
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "docker",
							MountPath: "/var/run/docker.sock",
						},
						{
							Name:      "cgroup",
							MountPath: "/sys/fs/cgroup",
						},
						// containerd client will need to access /var/data, /run/containerd and /run/runc
						{
							Name:      "vardata",
							MountPath: "/var/data",
						},
						{
							Name:      "runcontainerd",
							MountPath: "/run/containerd",
						},
						{
							Name:      "runrunc",
							MountPath: "/run/runc",
						},
						{
							Name:             "lxcfs",
							MountPath:        "/var/lib/lxc",
							MountPropagation: &prop,
						},
					},
					Ports: []corev1.ContainerPort{
						{
							Name:          "http",
							HostPort:      int32(o.AgentPort),
							ContainerPort: 10027,
						},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "docker",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/var/run/docker.sock",
						},
					},
				},
				{
					Name: "cgroup",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/sys/fs/cgroup",
						},
					},
				},
				{
					Name: "lxcfs",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/var/lib/lxc",
							Type: &directoryCreate,
						},
					},
				},
				{
					Name: "vardata",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/var/data",
						},
					},
				},
				{
					Name: "runcontainerd",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/run/containerd",
						},
					},
				},
				{
					Name: "runrunc",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/run/runc",
						},
					},
				},
			},
			RestartPolicy: corev1.RestartPolicyNever,
		},
	}
	fmt.Fprintf(o.Out, "Agent Pod info: [Name:%s, Namespace:%s, Image:%s, HostPort:%d, ContainerPort:%d]\n", agentPod.ObjectMeta.Name, agentPod.ObjectMeta.Namespace, agentPod.Spec.Containers[0].Image, agentPod.Spec.Containers[0].Ports[0].HostPort, agentPod.Spec.Containers[0].Ports[0].ContainerPort)
	return agentPod
}

func (o *DebugOptions) runPortForward(pod *corev1.Pod) error {
	if pod.Status.Phase != corev1.PodRunning {
		return fmt.Errorf("unable to forward port because pod is not running. Current status=%v", pod.Status.Phase)
	}
	o.wait.Add(1)
	go func() {
		defer o.wait.Done()
		req := o.RESTClient.Post().
			Resource("pods").
			Namespace(pod.Namespace).
			Name(pod.Name).
			SubResource("portforward")
		err := o.PortForwarder.ForwardPorts("POST", req.URL(), o)
		if err != nil {
			log.Printf("PortForwarded failed with %+v\r\n", err)
			log.Printf("Sending ready signal just in case the failure reason is that the port is already forwarded.\r\n")
			o.ReadyChannel <- struct{}{}
		}
		if o.Verbosity > 0 {
			fmt.Fprintln(o.Out, "end port-forward...")
		}
	}()
	return nil
}

type portForwarder interface {
	ForwardPorts(method string, url *url.URL, opts *DebugOptions) error
}

type defaultPortForwarder struct {
	genericclioptions.IOStreams
}

// ForwardPorts forward ports
func (f *defaultPortForwarder) ForwardPorts(method string, url *url.URL, opts *DebugOptions) error {
	transport, upgrader, err := spdy.RoundTripperFor(opts.Config)
	if err != nil {
		return err
	}
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, method, url)
	fw, err := portforward.New(dialer, opts.Ports, opts.StopChannel, opts.ReadyChannel, f.Out, f.ErrOut)
	if err != nil {
		return err
	}
	return fw.ForwardPorts()
}

// auth checks if current user has permission to create pods/exec subresource.
func (o *DebugOptions) auth(ctx context.Context, pod *corev1.Pod) error {
	sarClient := o.KubeCli.AuthorizationV1()
	sar := &authorizationv1.SelfSubjectAccessReview{
		Spec: authorizationv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authorizationv1.ResourceAttributes{
				Namespace:   pod.Namespace,
				Verb:        "create",
				Group:       "",
				Resource:    "pods",
				Subresource: "exec",
				Name:        "",
			},
		},
	}
	response, err := sarClient.SelfSubjectAccessReviews().Create(ctx, sar, v1.CreateOptions{})
	if err != nil {
		fmt.Fprintf(o.ErrOut, "Failed to create SelfSubjectAccessReview: %v \n", err)
		return err
	}
	if !response.Status.Allowed {
		denyReason := fmt.Sprintf("Current user has no permission to create pods/exec subresource in namespace:%s. Detail:", pod.Namespace)
		if len(response.Status.Reason) > 0 {
			denyReason = fmt.Sprintf("%s %v, ", denyReason, response.Status.Reason)
		}
		if len(response.Status.EvaluationError) > 0 {
			denyReason = fmt.Sprintf("%s %v", denyReason, response.Status.EvaluationError)
		}
		return fmt.Errorf(denyReason)
	}
	return nil
}

// delete the agent pod
func (o *DebugOptions) deleteAgent(ctx context.Context, agentPod *corev1.Pod) {
	// only with agentless flag we can delete the agent pod
	if !o.AgentLess {
		return
	}
	err := o.CoreClient.Pods(agentPod.Namespace).Delete(ctx, agentPod.Name, *v1.NewDeleteOptions(0))
	if err != nil {
		fmt.Fprintf(o.ErrOut, "failed to delete agent pod[Name:%s, Namespace: %s], consider manual deletion.\nerror msg: %v", agentPod.Name, agentPod.Namespace, err)
	}
}

// build the agent pod Resource Requirements
func (o *DebugOptions) buildAgentResourceRequirements() corev1.ResourceRequirements {
	return getResourceRequirements(getResourceList(o.AgentPodResource.CpuRequests, o.AgentPodResource.MemoryRequests), getResourceList(o.AgentPodResource.CpuLimits, o.AgentPodResource.MemoryLimits))
}

func getResourceList(cpu, memory string) corev1.ResourceList {
	// catch error in resource.MustParse
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("Parse Resource list error: %v\n", err)
		}
	}()
	res := corev1.ResourceList{}
	if cpu != "" {
		res[corev1.ResourceCPU] = resource.MustParse(cpu)
	}
	if memory != "" {
		res[corev1.ResourceMemory] = resource.MustParse(memory)
	}
	return res
}

func getResourceRequirements(requests, limits corev1.ResourceList) corev1.ResourceRequirements {
	res := corev1.ResourceRequirements{}
	res.Requests = requests
	res.Limits = limits
	return res
}
