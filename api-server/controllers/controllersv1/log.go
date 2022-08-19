package controllersv1

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/bentoml/yatai-schemas/schemasv1"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/common/consts"
)

type logMessageType string

const (
	logMessageTypeReplace logMessageType = "replace"
	logMessageTypeAppend  logMessageType = "append"
)

type logMessage struct {
	ReqId string         `json:"req_id"`
	Type  logMessageType `json:"type"`
	Items []string       `json:"items"`
}

type Tail struct {
	Finished      bool
	conn          *websocket.Conn
	closeCh       chan struct{}
	toClose       chan struct{}
	err           chan error
	namespace     string
	podNames      []string
	containerName string
	timestamps    bool

	enableLogName bool

	currentReqId string
	mu           sync.Mutex
}

type tailRequest struct {
	Id            string  `json:"id"`
	TailLines     *int64  `json:"tail_lines"`
	ContainerName *string `json:"container_name"`
	SinceTime     *time.Time
	Follow        bool
}

type wsTailRequest struct {
	schemasv1.WsReqSchema
	Payload *tailRequest `json:"payload"`
}

// NewTail creates new Tail object
func NewTail(conn *websocket.Conn, namespace string, podNames []string, containerName string, timestamps, enableLogName bool) *Tail {
	return &Tail{
		Finished:      false,
		conn:          conn,
		closeCh:       make(chan struct{}),
		toClose:       make(chan struct{}, 1),
		err:           make(chan error, 1),
		namespace:     namespace,
		podNames:      podNames,
		containerName: containerName,
		timestamps:    timestamps,
		enableLogName: enableLogName,
	}
}

func (t *Tail) Write(msg []byte) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.conn.WriteMessage(websocket.TextMessage, msg)
}

func (t *Tail) doClose(err error) {
	select {
	case t.toClose <- struct{}{}:
	default:
	}
	select {
	case t.err <- err:
	default:
	}
}

// Start starts Pod log streaming
func (t *Tail) Start(ctx context.Context, clientset *kubernetes.Clientset) error {
	go func() {
		<-t.toClose
		close(t.closeCh)
	}()

	go func() {
		select {
		case <-t.closeCh:
		case <-ctx.Done():
			t.doClose(nil)
		}
	}()

	reqCh := make(chan *tailRequest, 1)

	go func() {
		for {
			select {
			case <-t.closeCh:
				return
			default:
			}

			_, p, err := t.conn.ReadMessage()

			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					logrus.Errorf("[LOG] ws read failed: %q", err.Error())
				}
				t.doClose(err)
				return
			}

			req := wsTailRequest{}
			err = json.Unmarshal(p, &req)

			if err != nil {
				logrus.Errorf("marshal tail msg: %s", err.Error())
				continue
			}

			reqCh <- req.Payload
		}
	}()

	go func() {
		for {
			select {
			case <-t.closeCh:
				return
			case req := <-reqCh:
				t.currentReqId = req.Id

				logOptions := &v1.PodLogOptions{
					Container:  t.containerName,
					TailLines:  req.TailLines,
					Timestamps: t.timestamps,
				}

				if req.ContainerName != nil {
					logOptions.Container = *req.ContainerName
				}

				if req.SinceTime != nil {
					logOptions.SinceTime = &metav1.Time{
						Time: *req.SinceTime,
					}
				}

				now := time.Now()

				for _, podName := range t.podNames {
					podName := podName

					rs, err := clientset.CoreV1().Pods(t.namespace).GetLogs(podName, logOptions).Stream(ctx)
					if err != nil {
						t.doClose(errors.Wrapf(err, "get pod %s log failed", podName))
						return
					}

					err = func() error {
						defer rs.Close()
						buf := new(bytes.Buffer)
						_, err = io.Copy(buf, rs)
						if err != nil {
							return errors.Wrap(err, "error in copy information from podLogs to buf")
						}
						str := buf.String()
						lines := strings.Split(str, "\n")
						res := make([]string, 0, len(lines))
						for _, line := range lines {
							if t.enableLogName {
								line = fmt.Sprintf("[%s] [%s] %s", podName, t.containerName, line)
							}
							res = append(res, line)
						}
						msg := schemasv1.WsRespSchema{
							Type:    schemasv1.WsRespTypeSuccess,
							Message: "",
							Payload: &logMessage{
								ReqId: req.Id,
								Type:  logMessageTypeReplace,
								Items: res,
							},
						}
						msgStr, err := json.Marshal(&msg)
						if err != nil {
							return errors.Wrap(err, "error in marshal log message")
						}
						err = t.Write(msgStr)
						if err != nil {
							return errors.Wrap(err, "error in write log message")
						}
						if req.Follow {
							go func() {
								logOptions.Follow = true
								logOptions.TailLines = nil
								logOptions.SinceTime = &metav1.Time{
									Time: now,
								}
								rs, err := clientset.CoreV1().Pods(t.namespace).GetLogs(podName, logOptions).Stream(ctx)
								if err != nil {
									_, _ = fmt.Fprintln(os.Stderr, err)
									return
								}
								defer rs.Close()
								sc := bufio.NewScanner(rs)
								for sc.Scan() {
									select {
									case <-t.closeCh:
										return
									default:
									}

									if t.currentReqId != req.Id {
										break
									}

									content := sc.Text()
									if t.enableLogName {
										content = fmt.Sprintf("[%s] [%s] %s", podName, t.containerName, content)
									}
									msg := schemasv1.WsRespSchema{
										Type:    schemasv1.WsRespTypeSuccess,
										Message: "",
										Payload: &logMessage{
											ReqId: req.Id,
											Type:  logMessageTypeAppend,
											Items: []string{
												content,
											},
										},
									}
									msgStr, err := json.Marshal(&msg)
									if err != nil {
										t.doClose(errors.Wrap(err, "error in marshal log message"))
										return
									}
									err = t.Write(msgStr)
									if err != nil {
										t.doClose(errors.Wrap(err, "error in write log message"))
										return
									}
								}
							}()
						}
						return nil
					}()
					if err != nil {
						t.doClose(err)
						return
					}
				}
			}
		}
	}()

	<-t.closeCh
	select {
	case err := <-t.err:
		return err
	default:
		return nil
	}
}

type logController struct {
	baseController
}

var LogController = logController{}

func (c *logController) TailDeploymentPodLog(ctx *gin.Context, schema *GetDeploymentSchema) error {
	var err error

	ctx.Request.Header.Del("Origin")
	conn, err := wsUpgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		logrus.Errorf("ws connect failed: %q", err.Error())
		return err
	}
	defer conn.Close()

	defer func() {
		writeWsError(conn, err)
	}()

	deployment, err := schema.GetDeployment(ctx)
	if err != nil {
		return err
	}

	if err = DeploymentController.canView(ctx, deployment); err != nil {
		return err
	}

	cluster, err := schema.GetCluster(ctx)
	if err != nil {
		return err
	}

	cliset, _, err := services.ClusterService.GetKubeCliSet(ctx, cluster)
	if err != nil {
		return err
	}

	podName := ctx.Query("pod_name")
	var podNames []string
	containerName := ctx.Query("container_name")

	kubeNs := services.DeploymentService.GetKubeNamespace(deployment)

	if podName != "" {
		podNames = append(podNames, podName)
		podsCli := cliset.CoreV1().Pods(kubeNs)

		pod, err := podsCli.Get(ctx, podName, metav1.GetOptions{})
		if err != nil {
			return err
		}

		if pod.Labels[consts.KubeLabelYataiDeployment] != deployment.Name {
			return errors.Errorf("pod %s not in this deployment", podName)
		}

		if containerName == "" {
			containerName = pod.Status.ContainerStatuses[0].Name
		}
	}

	t := NewTail(conn, kubeNs, podNames, containerName, true, false)

	err = t.Start(ctx, cliset)
	return err
}

func (c *logController) TailClusterPodLog(ctx *gin.Context, schema *GetClusterSchema) error {
	var err error

	ctx.Request.Header.Del("Origin")
	conn, err := wsUpgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		logrus.Errorf("ws connect failed: %q", err.Error())
		return err
	}
	defer conn.Close()

	defer func() {
		writeWsError(conn, err)
	}()

	cluster, err := schema.GetCluster(ctx)
	if err != nil {
		logrus.Errorf("get cluster failed: %q", err.Error())
		return err
	}

	if err = ClusterController.canView(ctx, cluster); err != nil {
		logrus.Errorf("can not view cluster: %q", err.Error())
		return err
	}

	cliset, _, err := services.ClusterService.GetKubeCliSet(ctx, cluster)
	if err != nil {
		logrus.Errorf("get kube cli set failed: %q", err.Error())
		return err
	}

	podName := ctx.Query("pod_name")
	var podNames []string
	containerName := ctx.Query("container_name")

	kubeNs := ctx.Query("namespace")

	if podName != "" {
		podNames = append(podNames, podName)
		podsCli := cliset.CoreV1().Pods(kubeNs)

		pod, err := podsCli.Get(ctx, podName, metav1.GetOptions{})
		if err != nil {
			logrus.Errorf("get pod failed: %q", err.Error())
			return err
		}

		if containerName == "" {
			containerName = pod.Status.ContainerStatuses[0].Name
		}
	}

	t := NewTail(conn, kubeNs, podNames, containerName, true, false)

	err = t.Start(ctx, cliset)
	return err
}
