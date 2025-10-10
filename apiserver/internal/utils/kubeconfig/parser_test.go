package kubeconfig

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseKubeconfig_Valid(t *testing.T) {
	kubeconfigYAML := `
apiVersion: v1
kind: Config
clusters:
- name: my-cluster
  cluster:
    server: https://kubernetes.example.com:6443
    certificate-authority-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0t
contexts:
- name: my-context
  context:
    cluster: my-cluster
    user: my-user
users:
- name: my-user
  user:
    token: my-secret-token
`

	contexts, err := ParseKubeconfig(kubeconfigYAML)
	require.NoError(t, err)
	require.Len(t, contexts, 1)

	ctx := contexts[0]
	assert.Equal(t, "my-context", ctx.Name)
	assert.Equal(t, "my-cluster", ctx.ClusterName)
	assert.Equal(t, "https://kubernetes.example.com:6443", ctx.Server)
	assert.Equal(t, "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0t", ctx.CertificateAuthorityData)
	assert.Equal(t, "my-user", ctx.UserName)
	assert.Equal(t, "my-secret-token", ctx.Token)
}

func TestParseKubeconfig_MultipleContexts(t *testing.T) {
	kubeconfigYAML := `
apiVersion: v1
kind: Config
clusters:
- name: cluster1
  cluster:
    server: https://cluster1.example.com
- name: cluster2
  cluster:
    server: https://cluster2.example.com
contexts:
- name: context1
  context:
    cluster: cluster1
    user: user1
- name: context2
  context:
    cluster: cluster2
    user: user2
users:
- name: user1
  user:
    token: token1
- name: user2
  user:
    client-certificate-data: cert-data
    client-key-data: key-data
`

	contexts, err := ParseKubeconfig(kubeconfigYAML)
	require.NoError(t, err)
	require.Len(t, contexts, 2)

	// Check first context
	assert.Equal(t, "context1", contexts[0].Name)
	assert.Equal(t, "cluster1", contexts[0].ClusterName)
	assert.Equal(t, "https://cluster1.example.com", contexts[0].Server)
	assert.Equal(t, "user1", contexts[0].UserName)
	assert.Equal(t, "token1", contexts[0].Token)

	// Check second context
	assert.Equal(t, "context2", contexts[1].Name)
	assert.Equal(t, "cluster2", contexts[1].ClusterName)
	assert.Equal(t, "https://cluster2.example.com", contexts[1].Server)
	assert.Equal(t, "user2", contexts[1].UserName)
	assert.Equal(t, "cert-data", contexts[1].ClientCertificateData)
	assert.Equal(t, "key-data", contexts[1].ClientKeyData)
}

func TestParseKubeconfig_InvalidYAML(t *testing.T) {
	kubeconfigYAML := `
not valid yaml: {{{
`

	_, err := ParseKubeconfig(kubeconfigYAML)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse YAML")
}

func TestParseKubeconfig_MissingAPIVersion(t *testing.T) {
	kubeconfigYAML := `
kind: Config
clusters:
- name: my-cluster
  cluster:
    server: https://kubernetes.example.com
contexts:
- name: my-context
  context:
    cluster: my-cluster
    user: my-user
users:
- name: my-user
  user:
    token: token
`

	_, err := ParseKubeconfig(kubeconfigYAML)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "apiVersion")
}

func TestParseKubeconfig_InvalidKind(t *testing.T) {
	kubeconfigYAML := `
apiVersion: v1
kind: InvalidKind
clusters:
- name: my-cluster
  cluster:
    server: https://kubernetes.example.com
contexts:
- name: my-context
  context:
    cluster: my-cluster
    user: my-user
users:
- name: my-user
  user:
    token: token
`

	_, err := ParseKubeconfig(kubeconfigYAML)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "kind must be 'Config'")
}

func TestParseKubeconfig_MissingClusters(t *testing.T) {
	kubeconfigYAML := `
apiVersion: v1
kind: Config
contexts:
- name: my-context
  context:
    cluster: my-cluster
    user: my-user
users:
- name: my-user
  user:
    token: token
`

	_, err := ParseKubeconfig(kubeconfigYAML)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one cluster is required")
}

func TestParseKubeconfig_MissingUsers(t *testing.T) {
	kubeconfigYAML := `
apiVersion: v1
kind: Config
clusters:
- name: my-cluster
  cluster:
    server: https://kubernetes.example.com
contexts:
- name: my-context
  context:
    cluster: my-cluster
    user: my-user
`

	_, err := ParseKubeconfig(kubeconfigYAML)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one user is required")
}

func TestParseKubeconfig_MissingContexts(t *testing.T) {
	kubeconfigYAML := `
apiVersion: v1
kind: Config
clusters:
- name: my-cluster
  cluster:
    server: https://kubernetes.example.com
users:
- name: my-user
  user:
    token: token
`

	_, err := ParseKubeconfig(kubeconfigYAML)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one context is required")
}

func TestParseKubeconfig_MissingServer(t *testing.T) {
	kubeconfigYAML := `
apiVersion: v1
kind: Config
clusters:
- name: my-cluster
  cluster:
    certificate-authority-data: cert-data
contexts:
- name: my-context
  context:
    cluster: my-cluster
    user: my-user
users:
- name: my-user
  user:
    token: token
`

	_, err := ParseKubeconfig(kubeconfigYAML)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "server URL is required")
}

func TestParseKubeconfig_ClusterNotFound(t *testing.T) {
	kubeconfigYAML := `
apiVersion: v1
kind: Config
clusters:
- name: my-cluster
  cluster:
    server: https://kubernetes.example.com
contexts:
- name: my-context
  context:
    cluster: non-existent-cluster
    user: my-user
users:
- name: my-user
  user:
    token: token
`

	_, err := ParseKubeconfig(kubeconfigYAML)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cluster 'non-existent-cluster' not found")
}

func TestParseKubeconfig_UserNotFound(t *testing.T) {
	kubeconfigYAML := `
apiVersion: v1
kind: Config
clusters:
- name: my-cluster
  cluster:
    server: https://kubernetes.example.com
contexts:
- name: my-context
  context:
    cluster: my-cluster
    user: non-existent-user
users:
- name: my-user
  user:
    token: token
`

	_, err := ParseKubeconfig(kubeconfigYAML)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user 'non-existent-user' not found")
}

func TestParseKubeconfig_DuplicateClusterNames(t *testing.T) {
	kubeconfigYAML := `
apiVersion: v1
kind: Config
clusters:
- name: my-cluster
  cluster:
    server: https://kubernetes1.example.com
- name: my-cluster
  cluster:
    server: https://kubernetes2.example.com
contexts:
- name: my-context
  context:
    cluster: my-cluster
    user: my-user
users:
- name: my-user
  user:
    token: token
`

	_, err := ParseKubeconfig(kubeconfigYAML)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate cluster name")
}

func TestParseKubeconfig_DuplicateUserNames(t *testing.T) {
	kubeconfigYAML := `
apiVersion: v1
kind: Config
clusters:
- name: my-cluster
  cluster:
    server: https://kubernetes.example.com
contexts:
- name: my-context
  context:
    cluster: my-cluster
    user: my-user
users:
- name: my-user
  user:
    token: token1
- name: my-user
  user:
    token: token2
`

	_, err := ParseKubeconfig(kubeconfigYAML)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate user name")
}

func TestParseKubeconfig_DuplicateContextNames(t *testing.T) {
	kubeconfigYAML := `
apiVersion: v1
kind: Config
clusters:
- name: my-cluster
  cluster:
    server: https://kubernetes.example.com
contexts:
- name: my-context
  context:
    cluster: my-cluster
    user: my-user
- name: my-context
  context:
    cluster: my-cluster
    user: my-user
users:
- name: my-user
  user:
    token: token
`

	_, err := ParseKubeconfig(kubeconfigYAML)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate context name")
}

func TestParseKubeconfig_EmptyClusterName(t *testing.T) {
	kubeconfigYAML := `
apiVersion: v1
kind: Config
clusters:
- name: ""
  cluster:
    server: https://kubernetes.example.com
contexts:
- name: my-context
  context:
    cluster: ""
    user: my-user
users:
- name: my-user
  user:
    token: token
`

	_, err := ParseKubeconfig(kubeconfigYAML)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cluster name cannot be empty")
}

func TestParseKubeconfig_EmptyUserName(t *testing.T) {
	kubeconfigYAML := `
apiVersion: v1
kind: Config
clusters:
- name: my-cluster
  cluster:
    server: https://kubernetes.example.com
contexts:
- name: my-context
  context:
    cluster: my-cluster
    user: ""
users:
- name: ""
  user:
    token: token
`

	_, err := ParseKubeconfig(kubeconfigYAML)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user name cannot be empty")
}

func TestParseKubeconfig_EmptyContextName(t *testing.T) {
	kubeconfigYAML := `
apiVersion: v1
kind: Config
clusters:
- name: my-cluster
  cluster:
    server: https://kubernetes.example.com
contexts:
- name: ""
  context:
    cluster: my-cluster
    user: my-user
users:
- name: my-user
  user:
    token: token
`

	_, err := ParseKubeconfig(kubeconfigYAML)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context name cannot be empty")
}
