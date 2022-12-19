/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllersv1

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	watchtools "k8s.io/client-go/tools/watch"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
)

const (
	DebuggerImage = "quay.io/bentoml/bento-debugger:0.0.5"
)

// profileLegacy represents the legacy debugging profile which is backwards-compatible with 1.23 behavior.
func profileLegacy(pod *corev1.Pod, containerName string, target runtime.Object) error {
	switch target.(type) {
	case *corev1.Pod:
		// do nothing to the copied pod
		return nil
	case *corev1.Node:
		const volumeName = "host-root"
		pod.Spec.Volumes = append(pod.Spec.Volumes, corev1.Volume{
			Name: volumeName,
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{Path: "/"},
			},
		})

		for i := range pod.Spec.Containers {
			container := &pod.Spec.Containers[i]
			if container.Name != containerName {
				continue
			}
			container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
				MountPath: "/host",
				Name:      volumeName,
			})
		}

		pod.Spec.HostIPC = true
		pod.Spec.HostNetwork = true
		pod.Spec.HostPID = true
		return nil
	default:
		return fmt.Errorf("the %s profile doesn't support objects of type %T", ProfileLegacy, target)
	}
}

// ProfileLegacy represents the legacy debugging profile which is backwards-compatible with 1.23 behavior.
const ProfileLegacy = "legacy"

type ProfileApplier interface {
	// Apply applies the profile to the given container in the pod.
	Apply(pod *corev1.Pod, containerName string, target runtime.Object) error
}

// NewProfileApplier returns a new Options for the given profile name.
func NewProfileApplier(profile string) (ProfileApplier, error) {
	// nolint: gocritic
	switch profile {
	case ProfileLegacy:
		return applierFunc(profileLegacy), nil
	}

	return nil, fmt.Errorf("unknown profile: %s", profile)
}

// applierFunc is a function that applies a profile to a container in the pod.
type applierFunc func(pod *corev1.Pod, containerName string, target runtime.Object) error

func (f applierFunc) Apply(pod *corev1.Pod, containerName string, target runtime.Object) error {
	return f(pod, containerName, target)
}

// DebugOptions specify how to run debug container in a running pod
type DebugOptions struct {
	Args            []string
	ArgsOnly        bool
	Env             []corev1.EnvVar
	Image           string
	Interactive     bool
	Namespace       string
	PullPolicy      corev1.PullPolicy
	Quiet           bool
	TargetContainer string
	TTY             bool

	podClient corev1client.CoreV1Interface

	applier ProfileApplier
}

func (o *DebugOptions) Complete(podClient corev1client.CoreV1Interface) (err error) {
	o.podClient = podClient
	o.applier, err = NewProfileApplier(ProfileLegacy)
	if err != nil {
		return
	}
	return
}

func containerNameToRef(pod *corev1.Pod) map[string]*corev1.Container {
	names := map[string]*corev1.Container{}
	for i := range pod.Spec.Containers {
		ref := &pod.Spec.Containers[i]
		names[ref.Name] = ref
	}
	for i := range pod.Spec.InitContainers {
		ref := &pod.Spec.InitContainers[i]
		names[ref.Name] = ref
	}
	for i := range pod.Spec.EphemeralContainers {
		ref := (*corev1.Container)(&pod.Spec.EphemeralContainers[i].EphemeralContainerCommon)
		names[ref.Name] = ref
	}
	return names
}

func (o *DebugOptions) computeDebugContainerName(pod *corev1.Pod) string {
	cn, containerByName := "", containerNameToRef(pod)
	for len(cn) == 0 || (containerByName[cn] != nil) {
		cn = fmt.Sprintf("debugger-%s", nameSuffixFunc(5))
	}
	return cn
}

func (o *DebugOptions) generateDebugContainer(pod *corev1.Pod) (*corev1.Pod, *corev1.EphemeralContainer, error) {
	name := o.computeDebugContainerName(pod)
	var targetContainer *corev1.Container
	for _, c := range pod.Spec.Containers {
		c := c
		if c.Name == o.TargetContainer {
			targetContainer = &c
			break
		}
	}
	if targetContainer == nil {
		err := errors.Errorf("unable to find target container %s in pod %s", o.TargetContainer, pod.Name)
		return nil, nil, err
	}

	ec := &corev1.EphemeralContainer{
		EphemeralContainerCommon: corev1.EphemeralContainerCommon{
			Name:                     name,
			Env:                      o.Env,
			Image:                    o.Image,
			ImagePullPolicy:          o.PullPolicy,
			Stdin:                    o.Interactive,
			TerminationMessagePolicy: corev1.TerminationMessageReadFile,
			TTY:                      o.TTY,
			SecurityContext:          targetContainer.SecurityContext,
		},
		TargetContainerName: o.TargetContainer,
	}

	if o.ArgsOnly {
		ec.Args = o.Args
	} else {
		ec.Command = o.Args
	}

	copied := pod.DeepCopy()
	copied.Spec.EphemeralContainers = append(copied.Spec.EphemeralContainers, *ec)
	if err := o.applier.Apply(copied, name, copied); err != nil {
		return nil, nil, err
	}

	return copied, ec, nil
}

func (o *DebugOptions) debugByEphemeralContainer(ctx context.Context, pod *corev1.Pod) (*corev1.Pod, string, error) {
	podJS, err := json.Marshal(pod)
	if err != nil {
		err = errors.Wrap(err, "unable to marshal pod")
		return nil, "", err
	}

	debugPod, debugContainer, err := o.generateDebugContainer(pod)
	if err != nil {
		err = errors.Wrap(err, "unable to generate debug container")
		return nil, "", err
	}

	debugJS, err := json.Marshal(debugPod)
	if err != nil {
		err = errors.Wrap(err, "error creating JSON for debug pod")
		return nil, "", err
	}

	patch, err := strategicpatch.CreateTwoWayMergePatch(podJS, debugJS, pod)
	if err != nil {
		err = errors.Wrap(err, "error creating patch for debug pod")
		return nil, "", err
	}

	pods := o.podClient.Pods(pod.Namespace)
	result, err := pods.Patch(ctx, pod.Name, types.StrategicMergePatchType, patch, metav1.PatchOptions{}, "ephemeralcontainers")
	if err != nil {
		// The apiserver will return a 404 when the EphemeralContainers feature is disabled because the `/ephemeralcontainers` subresource
		// is missing. Unlike the 404 returned by a missing pod, the status details will be empty.
		// nolint: errorlint
		if serr, ok := err.(*k8serrors.StatusError); ok && serr.Status().Reason == metav1.StatusReasonNotFound && serr.ErrStatus.Details.Name == "" {
			return nil, "", fmt.Errorf("ephemeral containers are disabled for this cluster (error from server: %w).", err)
		}

		// The Kind used for the /ephemeralcontainers subresource changed in 1.22. When presented with an unexpected
		// Kind the api server will respond with a not-registered error. When this happens we can optimistically try
		// using the old API.
		if runtime.IsNotRegisteredError(err) {
			return o.debugByEphemeralContainerLegacy(ctx, pod, debugContainer)
		}

		return nil, "", err
	}

	return result, debugContainer.Name, nil
}

func (o *DebugOptions) debugByEphemeralContainerLegacy(ctx context.Context, pod *corev1.Pod, debugContainer *corev1.EphemeralContainer) (*corev1.Pod, string, error) {
	// We no longer have the v1.EphemeralContainers Kind since it was removed in 1.22, but
	// we can present a JSON 6902 patch that the api server will apply.
	patch, err := json.Marshal([]map[string]interface{}{{
		"op":    "add",
		"path":  "/ephemeralContainers/-",
		"value": debugContainer,
	}})
	if err != nil {
		return nil, "", fmt.Errorf("error creating JSON 6902 patch for old /ephemeralcontainers API: %w", err)
	}

	result := o.podClient.RESTClient().Patch(types.JSONPatchType).
		Namespace(pod.Namespace).
		Resource("pods").
		Name(pod.Name).
		SubResource("ephemeralcontainers").
		Body(patch).
		Do(ctx)
	if err := result.Error(); err != nil {
		return nil, "", err
	}

	newPod, err := o.podClient.Pods(pod.Namespace).Get(ctx, pod.Name, metav1.GetOptions{})
	if err != nil {
		return nil, "", err
	}

	return newPod, debugContainer.Name, nil
}

func (o *DebugOptions) LoadDebuggerContainer(ctx context.Context, pod *corev1.Pod) (*corev1.Pod, string, error) {
	pod, containerName, err := o.debugByEphemeralContainer(ctx, pod)
	if err != nil {
		return nil, "", err
	}
	pod, err = o.waitForContainer(ctx, o.Namespace, pod.Name, containerName)
	if err != nil {
		return nil, "", err
	}
	return pod, containerName, nil
}

func (o *DebugOptions) waitForContainer(ctx context.Context, ns, podName, containerName string) (*corev1.Pod, error) {
	// TODO: expose the timeout
	ctx, cancel := watchtools.ContextWithOptionalTimeout(ctx, 0*time.Second)
	defer cancel()

	fieldSelector := fields.OneTermEqualSelector("metadata.name", podName).String()
	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			options.FieldSelector = fieldSelector
			return o.podClient.Pods(ns).List(ctx, options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			options.FieldSelector = fieldSelector
			return o.podClient.Pods(ns).Watch(ctx, options)
		},
	}

	intr := NewInterruptHandler(nil, cancel)
	var result *corev1.Pod
	err := intr.Run(func() error {
		ev, err := watchtools.UntilWithSync(ctx, lw, &corev1.Pod{}, nil, func(ev watch.Event) (bool, error) {
			// nolint: gocritic, exhaustive
			switch ev.Type {
			case watch.Deleted:
				return false, k8serrors.NewNotFound(schema.GroupResource{Resource: "pods"}, "")
			}

			p, ok := ev.Object.(*corev1.Pod)
			if !ok {
				return false, fmt.Errorf("watch did not return a pod: %v", ev.Object)
			}

			s := getContainerStatusByName(p, containerName)
			if s == nil {
				return false, nil
			}
			if s.State.Running != nil || s.State.Terminated != nil {
				return true, nil
			}
			if !o.Quiet && s.State.Waiting != nil && s.State.Waiting.Message != "" {
				logrus.Warnf("Waiting for container %s: %s", containerName, s.State.Waiting.Message)
			}
			return false, nil
		})
		if ev != nil {
			result = ev.Object.(*corev1.Pod)
		}
		return err
	})

	return result, err
}

func getContainerStatusByName(pod *corev1.Pod, containerName string) *corev1.ContainerStatus {
	allContainerStatus := [][]corev1.ContainerStatus{pod.Status.InitContainerStatuses, pod.Status.ContainerStatuses, pod.Status.EphemeralContainerStatuses}
	for _, statusSlice := range allContainerStatus {
		for i := range statusSlice {
			if statusSlice[i].Name == containerName {
				return &statusSlice[i]
			}
		}
	}
	return nil
}
