package config

import (
	"testing"

	apiv1 "k8s.io/api/core/v1"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func TestDatabaseConfig(t *testing.T) {
	assert.Equal(t, "my-host", DatabaseConfig{Host: "my-host"}.GetHostname())
	assert.Equal(t, "my-host:1234", DatabaseConfig{Host: "my-host", Port: 1234}.GetHostname())
}

func TestContainerRuntimeExecutor(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		c := Config{ContainerRuntimeExecutor: "foo"}
		executor, err := c.GetContainerRuntimeExecutor(labels.Set{})
		assert.NoError(t, err)
		assert.Equal(t, "foo", executor)
	})
	t.Run("Error", func(t *testing.T) {
		c := Config{ContainerRuntimeExecutor: "foo", ContainerRuntimeExecutors: ContainerRuntimeExecutors{
			{Name: "bar", Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{"!": "!"},
			}},
		}}
		_, err := c.GetContainerRuntimeExecutor(labels.Set{})
		assert.Error(t, err)
	})
	t.Run("NoError", func(t *testing.T) {
		c := Config{ContainerRuntimeExecutor: "foo", ContainerRuntimeExecutors: ContainerRuntimeExecutors{
			{Name: "bar", Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{"baz": "qux"},
			}},
		}}
		executor, err := c.GetContainerRuntimeExecutor(labels.Set(map[string]string{"baz": "qux"}))
		assert.NoError(t, err)
		assert.Equal(t, "bar", executor)
	})
}

func TestEventSinkConfig(t *testing.T) {
	sinkFoo := SinkConfig{Name: "foo"}
	sinkBar := SinkConfig{Name: "bar"}
	sinks := make([]SinkConfig, 2)
	sinks = append(sinks, sinkFoo)
	sinks = append(sinks, sinkBar)

	eventSinksConfig := EventSinksConfig{
		Sinks: sinks,
	}

	assert.Equal(t, sinkFoo, eventSinksConfig.GetSinkConfig("foo"))
}

func TestRuleMatchEvent(t *testing.T) {
	rule := Rule{
		Kind: "Workflow",
	}
	annotations := map[string]string{}
	assert.True(t, rule.MatchesEvent(apiv1.EventTypeNormal, "WorkflowKind", "WorkflowSucceeded", "Workflow succeeded!", annotations))
}
