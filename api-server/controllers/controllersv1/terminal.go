package controllersv1

import (
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/bentoml/yatai/schemas/schemasv1"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/common/consts"
	"github.com/bentoml/yatai/common/utils"
	"github.com/bentoml/yatai/schemas/modelschemas"

	rbacv1 "k8s.io/api/rbac/v1"
	errors2 "k8s.io/apimachinery/pkg/api/errors"

	"k8s.io/client-go/tools/watch"
	"k8s.io/kubernetes/pkg/client/conditions"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/docker/docker/pkg/term"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubernetes/pkg/util/interrupt"
)

type terminalController struct {
	baseController
}

var TerminalController = terminalController{}

type xtermMessage struct {
	Type  string `json:"type"`
	Input string `json:"input"`
	Rows  uint16 `json:"rows"`
	Cols  uint16 `json:"cols"`
}

const (
	END_OF_TRANSMISSION = "\u0004"
	errorTimeOut        = "\033[31m30 min no input close\033[0m"
)

type WebTerminal struct {
	conn     *websocket.Conn
	sizeChan chan remotecommand.TerminalSize

	timeout   time.Duration
	readSize  chan int
	isTimeout bool
	err       error

	namespace     string
	podName       string
	containerName string

	recorder  *models.TerminalRecord
	closed    bool
	closeChan chan struct{}
}

func NewWebTerminal(ctx context.Context, conn *websocket.Conn, namespace, podName, containerName string, recorder *models.TerminalRecord) (*WebTerminal, error) {
	t := &WebTerminal{
		conn:     conn,
		sizeChan: make(chan remotecommand.TerminalSize),

		timeout:  time.Minute * 30,
		readSize: make(chan int, 1),

		namespace:     namespace,
		podName:       podName,
		containerName: containerName,

		recorder:  recorder,
		closeChan: make(chan struct{}),
	}

	go func() {
		for {
			t.watchRead()
			if t.err != nil {
				fmt.Println("term.watchRead break")
				break
			}
		}
	}()

	return t, nil
}

func (t *WebTerminal) ConnRead(buffer []byte, data string) (size int, err error) {
	if t.recorder != nil {
		err = services.TerminalRecordService.Append(context.Background(), t.recorder, modelschemas.RecordTypeInput, data)
		if err != nil {
			return
		}
	}
	size = copy(buffer, data)
	return
}

func (t *WebTerminal) ConnWrite(messageType int, data []byte) error {
	if t.recorder != nil {
		err := services.TerminalRecordService.Append(context.Background(), t.recorder, modelschemas.RecordTypeOutput, string(data))
		if err != nil {
			return err
		}
	}
	txt := b64.StdEncoding.EncodeToString(data)
	return t.conn.WriteMessage(messageType, []byte(txt))
}

func (t *WebTerminal) Read(buffer []byte) (size int, err error) {
	if t.closed {
		return
	}

	if t.isTimeout {
		_ = t.ConnWrite(websocket.TextMessage, []byte(errorTimeOut))
		return 0, errors.New(errorTimeOut)
	}

	defer func() {
		t.err = err
		t.readSize <- size
	}()

	mt, p, err := t.conn.ReadMessage()

	if err != nil {
		size, _ = t.Close(buffer)
		return
	}

	if mt == websocket.CloseMessage || mt == -1 {
		size, err = t.Close(buffer)
		return
	}

	message := xtermMessage{}
	err = json.Unmarshal(p, &message)

	if err != nil {
		size, err = t.Close(buffer)
		return
	}

	switch message.Type {
	case "input":
		size, err = t.ConnRead(buffer, message.Input)
	case "resize":
		terminalSize := remotecommand.TerminalSize{
			Width:  message.Cols,
			Height: message.Rows,
		}
		t.sizeChan <- terminalSize
	default:
		// ignore
	}

	return
}

func (t *WebTerminal) watchRead() {
	tf := time.After(t.timeout)
	select {
	case <-tf:
		t.isTimeout = true
	case <-t.readSize:
		return
	}
}

func (t *WebTerminal) Write(b []byte) (size int, err error) {
	size = len(b)

	err = t.ConnWrite(websocket.TextMessage, b)

	return
}

func (t *WebTerminal) Next() *remotecommand.TerminalSize {
	select {
	case size := <-t.sizeChan:
		if t.recorder != nil {
			t.recorder.Meta.Width = size.Width
			t.recorder.Meta.Height = size.Height
		}
		return &size
	case <-t.closeChan:
		return nil
	}
}

func (t *WebTerminal) Close(buffer []byte) (size int, err error) {
	if t.closed {
		return
	}
	close(t.closeChan)
	t.closed = true
	if t.recorder != nil {
		err = services.TerminalRecordService.SaveContent(context.Background(), t.recorder)
		if err != nil {
			return
		}
	}
	size = copy(buffer, END_OF_TRANSMISSION)
	return
}

// Safe invokes the provided function and will attempt to ensure that when the
// function returns (or a termination signal is sent) that the terminal state
// is reset to the condition it was in prior to the function being invoked. If
// t.Raw is true the terminal will be put into raw mode prior to calling the function.
// If the input file descriptor is not a TTY and TryDev is true, the /dev/tty file
// will be opened (if available).
func (t *WebTerminal) Safe(fn func() error) error {
	inFd, _ := term.GetFdInfo(t.conn)

	state, err := term.SaveState(inFd)
	if err != nil {
		return err
	}

	return interrupt.Chain(nil, func() {
		_ = term.RestoreTerminal(inFd, state)
	}).Run(fn)
}

func (t *WebTerminal) HandleDebug(ctx context.Context, cliset *kubernetes.Clientset, restConfig *rest.Config, fork bool) error {
	o := NewDebugOptions(t, cliset, restConfig, false, fork, consts.YataiDebugImg)
	return o.Run()
}

func (t *WebTerminal) Handle(ctx context.Context, cliset *kubernetes.Clientset, restConfig *rest.Config, cmd []string) error {
	f := func() error {
		sshReq := cliset.CoreV1().RESTClient().Post().
			Resource("pods").
			Name(t.podName).
			Namespace(t.namespace).
			SubResource("exec").
			VersionedParams(&corev1.PodExecOptions{
				Container: t.containerName,
				Command:   cmd,
				Stdin:     true,
				Stdout:    true,
				Stderr:    true,
				TTY:       true,
			}, scheme.ParameterCodec)

		executor, err := remotecommand.NewSPDYExecutor(restConfig, http.MethodPost, sshReq.URL())
		if err != nil {
			return err
		}

		logrus.Info("connecting to pod...")

		return executor.Stream(remotecommand.StreamOptions{
			Stdin:             t,
			Stdout:            t,
			Stderr:            t,
			TerminalSizeQueue: t,
			Tty:               true,
		})
	}

	return t.Safe(f)
}

func (c *terminalController) GetDeploymentPodTerminal(ctx *gin.Context, schema *GetDeploymentSchema) error {
	var err error

	ctx.Request.Header.Del("Origin")
	conn, err := wsUpgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		logrus.Errorf("ws connect failed: %q", err.Error())
		return err
	}
	defer conn.Close()

	defer func() {
		if err != nil {
			msg := schemasv1.WsRespSchema{
				Type:    schemasv1.WsRespTypeError,
				Message: err.Error(),
				Payload: nil,
			}
			_ = conn.WriteJSON(&msg)
		}
	}()

	deployment, err := schema.GetDeployment(ctx)
	if err != nil {
		return err
	}

	if err = DeploymentController.canUpdate(ctx, deployment); err != nil {
		return err
	}

	cluster, err := schema.GetCluster(ctx)
	if err != nil {
		return err
	}

	cliset, restConfig, err := services.ClusterService.GetKubeCliSet(ctx, cluster)
	if err != nil {
		return err
	}

	podName := ctx.Query("pod_name")
	containerName := ctx.Query("container_name")

	kubeNs := services.DeploymentService.GetKubeNamespace(deployment)

	if podName != "" {
		podsCli := cliset.CoreV1().Pods(kubeNs)

		pod, err := podsCli.Get(ctx, podName, metav1.GetOptions{})
		if err != nil {
			return err
		}

		if pod.Labels[consts.KubeLabelYataiDeployment] != deployment.Name {
			return errors.Errorf("pod %s not in this deployment %s", podName, deployment.Name)
		}

		if containerName == "" {
			containerName = pod.Spec.Containers[0].Name
		}
	}

	debug := ctx.Query("debug")
	fork := ctx.Query("fork")

	currentUser, err := services.GetCurrentUser(ctx)
	if err != nil {
		return err
	}

	cmd := []string{"sh", "-c", "bash || sh"}

	recorder, err := services.TerminalRecordService.Create(ctx, services.CreateTerminalRecordOption{
		CreatorId:      currentUser.ID,
		OrganizationId: utils.UintPtr(cluster.OrganizationId),
		ClusterId:      utils.UintPtr(cluster.ID),
		DeploymentId:   utils.UintPtr(deployment.ID),
		Resource:       deployment,
		Shell:          "/bin/bash",
		PodName:        podName,
		ContainerName:  containerName,
	})
	if err != nil {
		return err
	}

	t, err := NewWebTerminal(ctx, conn, kubeNs, podName, containerName, recorder)
	if err != nil {
		return err
	}

	if debug == "1" {
		return t.HandleDebug(ctx, cliset, restConfig, fork == "1")
	}

	return t.Handle(ctx, cliset, restConfig, cmd)
}

func (c *terminalController) GetClusterPodTerminal(ctx *gin.Context, schema *GetClusterSchema) error {
	var err error

	ctx.Request.Header.Del("Origin")
	conn, err := wsUpgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		logrus.Errorf("ws connect failed: %q", err.Error())
		return err
	}
	defer conn.Close()

	defer func() {
		if err != nil {
			msg := schemasv1.WsRespSchema{
				Type:    schemasv1.WsRespTypeError,
				Message: err.Error(),
				Payload: nil,
			}
			_ = conn.WriteJSON(&msg)
		}
	}()

	cluster, err := schema.GetCluster(ctx)
	if err != nil {
		return err
	}

	if err = ClusterController.canUpdate(ctx, cluster); err != nil {
		return err
	}

	cliset, restConfig, err := services.ClusterService.GetKubeCliSet(ctx, cluster)
	if err != nil {
		return err
	}

	podName := ctx.Query("pod_name")
	containerName := ctx.Query("container_name")

	kubeNs := ctx.Query("namespace")

	if podName != "" {
		podsCli := cliset.CoreV1().Pods(kubeNs)

		var pod *corev1.Pod
		pod, err = podsCli.Get(ctx, podName, metav1.GetOptions{})
		if err != nil {
			return err
		}

		if containerName == "" {
			containerName = pod.Spec.Containers[0].Name
		}
	}

	debug := ctx.Query("debug")
	fork := ctx.Query("fork")

	currentUser, err := services.GetCurrentUser(ctx)
	if err != nil {
		return err
	}

	cmd := []string{"sh", "-c", "bash || sh"}

	recorder, err := services.TerminalRecordService.Create(ctx, services.CreateTerminalRecordOption{
		CreatorId:      currentUser.ID,
		OrganizationId: utils.UintPtr(cluster.OrganizationId),
		ClusterId:      utils.UintPtr(cluster.ID),
		Resource:       cluster,
		Shell:          "/bin/bash",
		PodName:        podName,
		ContainerName:  containerName,
	})
	if err != nil {
		return err
	}

	t, err := NewWebTerminal(ctx, conn, kubeNs, podName, containerName, recorder)
	if err != nil {
		return err
	}

	if debug == "1" {
		return t.HandleDebug(ctx, cliset, restConfig, fork == "1")
	}

	return t.Handle(ctx, cliset, restConfig, cmd)
}

// nolint:unused,deadcode
func launchKubectlPod(ctx context.Context, cli *kubernetes.Clientset, userName string, logCh chan<- string) (*corev1.Pod, error) {
	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: consts.YataiKubectlNamespace,
			Name:      userName,
		},
	}
	serviceAccountCli := cli.CoreV1().ServiceAccounts(consts.YataiKubectlNamespace)
	_, err := serviceAccountCli.Get(ctx, serviceAccount.Name, metav1.GetOptions{})
	if errors2.IsNotFound(err) {
		serviceAccount, err = serviceAccountCli.Create(ctx, serviceAccount, metav1.CreateOptions{})
		if err != nil {
			return nil, errors.Wrap(err, "create service account")
		}
	} else if err != nil {
		return nil, errors.Wrap(err, "get service account")
	}
	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "mcd-cluster-admin",
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"*"},
				Resources: []string{"*"},
				Verbs:     []string{"*"},
			},
			{
				NonResourceURLs: []string{"*"},
				Verbs:           []string{"*"},
			},
		},
	}
	clusterRoleCli := cli.RbacV1().ClusterRoles()
	_, err = clusterRoleCli.Get(ctx, clusterRole.Name, metav1.GetOptions{})
	if errors2.IsNotFound(err) {
		clusterRole, err = clusterRoleCli.Create(ctx, clusterRole, metav1.CreateOptions{})
		if err != nil {
			return nil, errors.Wrap(err, "create cluster role")
		}
	} else if err != nil {
		return nil, errors.Wrap(err, "get cluster role")
	}
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("system:mcd-cluster-admin-%s", userName),
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     clusterRole.Name,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      serviceAccount.Name,
				Namespace: serviceAccount.Namespace,
			},
		},
	}
	clusterRoleBindingCli := cli.RbacV1().ClusterRoleBindings()
	_, err = clusterRoleBindingCli.Get(ctx, clusterRoleBinding.Name, metav1.GetOptions{})
	if errors2.IsNotFound(err) {
		_, err = clusterRoleBindingCli.Create(ctx, clusterRoleBinding, metav1.CreateOptions{})
		if err != nil {
			return nil, errors.Wrap(err, "create cluster role binding")
		}
	} else if err != nil {
		return nil, errors.Wrap(err, "get cluster role binding")
	}
	podName := fmt.Sprintf("mcd-kubectl-%s", userName)
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: consts.YataiKubectlNamespace,
			Labels: map[string]string{
				consts.KubeLabelMcdKubectl: consts.KubeLabelTrue,
				consts.KubeLabelMcdUser:    userName,
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:            consts.YataiKubectlContainerName,
					Image:           consts.YataiKubectlImage,
					ImagePullPolicy: corev1.PullAlways,
					Command:         []string{"sleep", "infinity"},
					LivenessProbe: &corev1.Probe{
						Handler: corev1.Handler{
							Exec: &corev1.ExecAction{
								Command: []string{"echo", "ok"},
							},
						},
					},
				},
			},
			RestartPolicy:      corev1.RestartPolicyNever,
			ServiceAccountName: serviceAccount.Name,
		},
	}

	logCh <- fmt.Sprintf("Launch pod %s...", pod.Name)

	return launchPod(ctx, cli, pod)
}

func launchPod(ctx context.Context, cli *kubernetes.Clientset, pod *corev1.Pod) (*corev1.Pod, error) {
	podCli := cli.CoreV1().Pods(pod.Namespace)
	_, err := podCli.Get(ctx, pod.Name, metav1.GetOptions{})
	// nolint: gocritic
	if errors2.IsNotFound(err) {
		pod, err = podCli.Create(ctx, pod, metav1.CreateOptions{})
		if err != nil {
			return pod, err
		}
	} else if err != nil {
		return nil, errors.Wrap(err, "get pod")
	} else {
		return pod, nil
	}

	watcher, err := podCli.Watch(ctx, metav1.SingleObject(pod.ObjectMeta))
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	event, err := watch.UntilWithoutRetry(ctx, watcher, conditions.PodRunning)
	if err != nil {
		return nil, err
	}
	pod = event.Object.(*corev1.Pod)
	return pod, nil
}

// nolint:unused,deadcode
func deletePod(ctx context.Context, cli *kubernetes.Clientset, pod *corev1.Pod) (*corev1.Pod, error) {
	err := cli.CoreV1().Pods(pod.Namespace).Delete(ctx, pod.Name, metav1.DeleteOptions{})
	if err != nil {
		return pod, err
	}
	return pod, nil
}
