package config

import (
	"fmt"
	"math"
	"regexp"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	wfv1 "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
)

var EmptyConfigFunc = func() interface{} { return &Config{} }

type ResourceRateLimit struct {
	Limit float64 `json:"limit"`
	Burst int     `json:"burst"`
}

// Config contain the configuration settings for the workflow controller
type Config struct {

	// NodeEvents configures how node events are omitted
	NodeEvents NodeEvents `json:"nodeEvents,omitempty"`

	// ExecutorImage is the image name of the executor to use when running pods
	// DEPRECATED: use --executor-image flag to workflow-controller instead
	ExecutorImage string `json:"executorImage,omitempty"`

	// ExecutorImagePullPolicy is the imagePullPolicy of the executor to use when running pods
	// DEPRECATED: use `executor.imagePullPolicy` in configmap instead
	ExecutorImagePullPolicy string `json:"executorImagePullPolicy,omitempty"`

	// Executor holds container customizations for the executor to use when running pods
	Executor *apiv1.Container `json:"executor,omitempty"`

	// ExecutorResources specifies the resource requirements that will be used for the executor sidecar
	// DEPRECATED: use `executor.resources` in configmap instead
	ExecutorResources *apiv1.ResourceRequirements `json:"executorResources,omitempty"`

	// MainContainer holds container customization for the main container
	MainContainer *apiv1.Container `json:"mainContainer,omitempty"`

	// KubeConfig specifies a kube config file for the wait & init containers
	KubeConfig *KubeConfig `json:"kubeConfig,omitempty"`

	// ContainerRuntimeExecutor specifies the container runtime interface to use, default is docker
	ContainerRuntimeExecutor string `json:"containerRuntimeExecutor,omitempty"`

	ContainerRuntimeExecutors ContainerRuntimeExecutors `json:"containerRuntimeExecutors,omitempty"`

	// KubeletPort is needed when using the kubelet containerRuntimeExecutor, default to 10250
	KubeletPort int `json:"kubeletPort,omitempty"`

	// KubeletInsecure disable the TLS verification of the kubelet containerRuntimeExecutor, default to false
	KubeletInsecure bool `json:"kubeletInsecure,omitempty"`

	// ArtifactRepository contains the default location of an artifact repository for container artifacts
	ArtifactRepository wfv1.ArtifactRepository `json:"artifactRepository,omitempty"`

	// Namespace is a label selector filter to limit the controller's watch to a specific namespace
	// DEPRECATED: support will be remove in a future release
	Namespace string `json:"namespace,omitempty"`

	// InstanceID is a label selector to limit the controller's watch to a specific instance. It
	// contains an arbitrary value that is carried forward into its pod labels, under the key
	// workflows.argoproj.io/controller-instanceid, for the purposes of workflow segregation. This
	// enables a controller to only receive workflow and pod events that it is interested about,
	// in order to support multiple controllers in a single cluster, and ultimately allows the
	// controller itself to be bundled as part of a higher level application. If omitted, the
	// controller watches workflows and pods that *are not* labeled with an instance id.
	InstanceID string `json:"instanceID,omitempty"`

	// MetricsConfig specifies configuration for metrics emission. Metrics are enabled and emitted on localhost:9090/metrics
	// by default.
	MetricsConfig MetricsConfig `json:"metricsConfig,omitempty"`

	// TelemetryConfig specifies configuration for telemetry emission. Telemetry is enabled and emitted in the same endpoint
	// as metrics by default, but can be overridden using this config.
	TelemetryConfig MetricsConfig `json:"telemetryConfig,omitempty"`

	// Parallelism limits the max total parallel workflows that can execute at the same time
	Parallelism int `json:"parallelism,omitempty"`

	// NamespaceParallelism limits the max workflows that can execute at the same time in a namespace
	NamespaceParallelism int `json:"namespaceParallelism,omitempty"`

	// ResourceRateLimit limits the rate at which pods are created
	ResourceRateLimit *ResourceRateLimit `json:"resourceRateLimit,omitempty"`

	// Persistence contains the workflow persistence DB configuration
	Persistence *PersistConfig `json:"persistence,omitempty"`

	// Links to related apps.
	Links []*wfv1.Link `json:"links,omitempty"`

	// Config customized Docker Sock path
	DockerSockPath string `json:"dockerSockPath,omitempty"`

	// WorkflowDefaults are values that will apply to all Workflows from this controller, unless overridden on the Workflow-level
	WorkflowDefaults *wfv1.Workflow `json:"workflowDefaults,omitempty"`

	// PodSpecLogStrategy enables the logging of podspec on controller log.
	PodSpecLogStrategy PodSpecLogStrategy `json:"podSpecLogStrategy,omitempty"`

	// PodGCGracePeriodSeconds specifies the duration in seconds before the pods in the GC queue get deleted.
	// Value must be non-negative integer. A zero value indicates that the pods will be deleted immediately
	// as soon as they arrived in the pod GC queue.
	// Defaults to 30 seconds.
	PodGCGracePeriodSeconds *int64 `json:"podGCGracePeriodSeconds,omitempty"`

	// WorkflowRestrictions restricts the controller to executing Workflows that meet certain restrictions
	WorkflowRestrictions *WorkflowRestrictions `json:"workflowRestrictions,omitempty"`

	// Adding configurable initial delay (for K8S clusters with mutating webhooks) to prevent workflow getting modified by MWC.
	InitialDelay metav1.Duration `json:"initialDelay,omitempty"`

	// The command/args for each image, needed when the command is not specified and the emissary executor is used.
	// https://argoproj.github.io/argo-workflows/workflow-executors/#emissary-emissary
	Images map[string]Image `json:"images,omitempty"`

	// The event sinks and rules to send controller events to them
	EventSinks EventSinksConfig `json:"eventSinks,omitempty"`
}

func (c Config) GetContainerRuntimeExecutor(labels labels.Labels) (string, error) {
	name, err := c.ContainerRuntimeExecutors.Select(labels)
	if err != nil {
		return "", err
	}
	if name != "" {
		return name, nil
	}
	return c.ContainerRuntimeExecutor, nil
}

func (c Config) GetResourceRateLimit() ResourceRateLimit {
	if c.ResourceRateLimit != nil {
		return *c.ResourceRateLimit
	}
	return ResourceRateLimit{
		Limit: math.MaxFloat32,
		Burst: math.MaxInt32,
	}
}

// PodSpecLogStrategy contains the configuration for logging the pod spec in controller log for debugging purpose
type PodSpecLogStrategy struct {
	FailedPod bool `json:"failedPod,omitempty"`
	AllPods   bool `json:"allPods,omitempty"`
}

// KubeConfig is used for wait & init sidecar containers to communicate with a k8s apiserver by a outofcluster method,
// it is used when the workflow controller is in a different cluster with the workflow workloads
type KubeConfig struct {
	// SecretName of the kubeconfig secret
	// may not be empty if kuebConfig specified
	SecretName string `json:"secretName"`
	// SecretKey of the kubeconfig in the secret
	// may not be empty if kubeConfig specified
	SecretKey string `json:"secretKey"`
	// VolumeName of kubeconfig, default to 'kubeconfig'
	VolumeName string `json:"volumeName,omitempty"`
	// MountPath of the kubeconfig secret, default to '/kube/config'
	MountPath string `json:"mountPath,omitempty"`
}

type PersistConfig struct {
	NodeStatusOffload bool `json:"nodeStatusOffLoad,omitempty"`
	// Archive workflows to persistence.
	Archive bool `json:"archive,omitempty"`
	// ArchivelabelSelector holds LabelSelector to determine workflow persistence.
	ArchiveLabelSelector *metav1.LabelSelector `json:"archiveLabelSelector,omitempty"`
	// in days
	ArchiveTTL     TTL               `json:"archiveTTL,omitempty"`
	ClusterName    string            `json:"clusterName,omitempty"`
	ConnectionPool *ConnectionPool   `json:"connectionPool,omitempty"`
	PostgreSQL     *PostgreSQLConfig `json:"postgresql,omitempty"`
	MySQL          *MySQLConfig      `json:"mysql,omitempty"`
	SkipMigration  bool              `json:"skipMigration,omitempty"`
}

func (c PersistConfig) GetArchiveLabelSelector() (labels.Selector, error) {
	if c.ArchiveLabelSelector == nil {
		return labels.Everything(), nil
	}
	return metav1.LabelSelectorAsSelector(c.ArchiveLabelSelector)
}

func (c PersistConfig) GetClusterName() string {
	if c.ClusterName != "" {
		return c.ClusterName
	}
	return "default"
}

type ConnectionPool struct {
	MaxIdleConns    int `json:"maxIdleConns,omitempty"`
	MaxOpenConns    int `json:"maxOpenConns,omitempty"`
	ConnMaxLifetime TTL `json:"connMaxLifetime,omitempty"`
}

type DatabaseConfig struct {
	Host           string                  `json:"host"`
	Port           int                     `json:"port,omitempty"`
	Database       string                  `json:"database"`
	TableName      string                  `json:"tableName,omitempty"`
	UsernameSecret apiv1.SecretKeySelector `json:"userNameSecret,omitempty"`
	PasswordSecret apiv1.SecretKeySelector `json:"passwordSecret,omitempty"`
}

func (c DatabaseConfig) GetHostname() string {
	if c.Port == 0 {
		return c.Host
	}
	return fmt.Sprintf("%s:%v", c.Host, c.Port)
}

type PostgreSQLConfig struct {
	DatabaseConfig
	SSL     bool   `json:"ssl,omitempty"`
	SSLMode string `json:"sslMode,omitempty"`
}

type MySQLConfig struct {
	DatabaseConfig
	Options map[string]string `json:"options,omitempty"`
}

// MetricsConfig defines a config for a metrics server
type MetricsConfig struct {
	// Enabled controls metric emission. Default is true, set "enabled: false" to turn off
	Enabled *bool `json:"enabled,omitempty"`
	// DisableLegacy turns off legacy metrics
	// DEPRECATED: Legacy metrics are now removed, this field is ignored
	DisableLegacy bool `json:"disableLegacy,omitempty"`
	// MetricsTTL sets how often custom metrics are cleared from memory
	MetricsTTL TTL `json:"metricsTTL,omitempty"`
	// Path is the path where metrics are emitted. Must start with a "/". Default is "/metrics"
	Path string `json:"path,omitempty"`
	// Port is the port where metrics are emitted. Default is "9090"
	Port int `json:"port,omitempty"`
	// IgnoreErrors is a flag that instructs prometheus to ignore metric emission errors
	IgnoreErrors bool `json:"ignoreErrors,omitempty"`
}

type WorkflowRestrictions struct {
	TemplateReferencing TemplateReferencing `json:"templateReferencing,omitempty"`
}

type TemplateReferencing string

const (
	TemplateReferencingStrict TemplateReferencing = "Strict"
	TemplateReferencingSecure TemplateReferencing = "Secure"
)

func (req *WorkflowRestrictions) MustUseReference() bool {
	if req == nil {
		return false
	}
	return req.TemplateReferencing == TemplateReferencingStrict || req.TemplateReferencing == TemplateReferencingSecure
}

func (req *WorkflowRestrictions) MustNotChangeSpec() bool {
	if req == nil {
		return false
	}
	return req.TemplateReferencing == TemplateReferencingSecure
}

type EventSinksConfig struct {
	Route *EventRoute  `json:"route,omitempty"`
	Sinks []SinkConfig `json:"sinks,omitempty"`
}

func (s *EventSinksConfig) GetSinkConfig(name string) SinkConfig {
	for _, sink := range s.Sinks {
		if sink.Name == name {
			return sink
		}
	}
	return SinkConfig{}
}

type EventRoute struct {
	Drop   []Rule       `json:"drop,omitempty"`
	Match  []Rule       `json:"match,omitempty"`
	Routes []EventRoute `json:"routes,omitempty"`
}

type Rule struct {
	Sink string `json:"sink"`
	// Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	Message     string            `json:"message,omitempty"`
	Kind        string            `json:"kind,omitempty"`
	Namespace   string            `json:"namespace,omitempty"`
	Reason      string            `json:"reason,omitempty"`
	Type        string            `json:"type,omitempty"`
}

// matchString is a method to clean the code. Error handling is omitted here because these
// rules are validated before use. According to regexp.MatchString, the only way it fails its
// that the pattern does not compile.
func matchString(pattern, s string) bool {
	matched, _ := regexp.MatchString(pattern, s)
	return matched
}

// MatchesEvent compares the rule to an event and returns a boolean value to indicate
// whether the event is compatible with the rule. All fields are compared as regular expressions
// so the user must keep that in mind while writing rules.
func (r *Rule) MatchesEvent(eventtype, kind, reason, message string, annotations map[string]string) bool {
	// These rules are just basic comparison rules, if one of them fails, it means the event does not match the rule
	rules := [][2]string{
		{r.Type, eventtype},
		{r.Kind, kind},
		{r.Reason, reason},
		{r.Message, message},
	}

	for _, v := range rules {
		rule := v[0]
		value := v[1]
		if rule != "" {
			matches := matchString(rule, value)
			if !matches {
				return false
			}
		}
	}

	// Annotations are also mutually exclusive, they all need to be present
	if r.Annotations != nil && len(r.Annotations) > 0 {
		for k, v := range r.Annotations {
			if val, ok := annotations[k]; !ok {
				return false
			} else {
				matches := matchString(v, val)
				if !matches {
					return false
				}
			}
		}
	}

	// If it failed every step, it must match because our matchers are limiting
	return true
}

func (r *Rule) MatchNamespace(namespace string) bool {
	return r.Namespace == namespace
}

type SinkConfig struct {
	Name    string         `json:"name"`
	Webhook *WebhookConfig `json:"webhook"`
}

type WebhookConfig struct {
	Endpoint string                 `json:"endpoint"`
	Layout   map[string]interface{} `json:"layout,omitempty"`
	Headers  map[string]string      `json:"headers,omitempty"`
}
