package kubeconfig

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// KubeconfigFile represents the structure of a kubeconfig YAML file
type KubeconfigFile struct {
	APIVersion     string             `yaml:"apiVersion"`
	Kind           string             `yaml:"kind"`
	Clusters       []NamedCluster     `yaml:"clusters"`
	Contexts       []NamedContext     `yaml:"contexts"`
	Users          []NamedUser        `yaml:"users"`
	CurrentContext string             `yaml:"current-context"`
	Preferences    map[string]interface{} `yaml:"preferences,omitempty"`
}

// NamedCluster represents a named cluster in kubeconfig
type NamedCluster struct {
	Name    string  `yaml:"name"`
	Cluster Cluster `yaml:"cluster"`
}

// Cluster represents cluster information
type Cluster struct {
	Server                   string `yaml:"server"`
	CertificateAuthorityData string `yaml:"certificate-authority-data,omitempty"`
	CertificateAuthority     string `yaml:"certificate-authority,omitempty"`
	InsecureSkipTLSVerify    bool   `yaml:"insecure-skip-tls-verify,omitempty"`
}

// NamedContext represents a named context in kubeconfig
type NamedContext struct {
	Name    string  `yaml:"name"`
	Context Context `yaml:"context"`
}

// Context represents context information
type Context struct {
	Cluster   string `yaml:"cluster"`
	User      string `yaml:"user"`
	Namespace string `yaml:"namespace,omitempty"`
}

// NamedUser represents a named user in kubeconfig
type NamedUser struct {
	Name string `yaml:"name"`
	User User   `yaml:"user"`
}

// User represents user authentication information
type User struct {
	Token                 string `yaml:"token,omitempty"`
	ClientCertificateData string `yaml:"client-certificate-data,omitempty"`
	ClientKeyData         string `yaml:"client-key-data,omitempty"`
	ClientCertificate     string `yaml:"client-certificate,omitempty"`
	ClientKey             string `yaml:"client-key,omitempty"`
	Username              string `yaml:"username,omitempty"`
	Password              string `yaml:"password,omitempty"`
	AuthProvider          map[string]interface{} `yaml:"auth-provider,omitempty"`
	Exec                  map[string]interface{} `yaml:"exec,omitempty"`
}

// ParsedContext represents a parsed and validated context with all necessary information
type ParsedContext struct {
	Name                     string
	ClusterName              string
	Server                   string
	CertificateAuthorityData string
	UserName                 string
	Token                    string
	ClientCertificateData    string
	ClientKeyData            string
}

// ValidationError represents a validation error with details
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ParseKubeconfig parses a kubeconfig YAML string and returns parsed contexts
func ParseKubeconfig(kubeconfigYAML string) ([]ParsedContext, error) {
	var config KubeconfigFile
	if err := yaml.Unmarshal([]byte(kubeconfigYAML), &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate basic structure
	if err := validateKubeconfig(&config); err != nil {
		return nil, err
	}

	// Build maps for quick lookup
	clusterMap := make(map[string]Cluster)
	for _, nc := range config.Clusters {
		clusterMap[nc.Name] = nc.Cluster
	}

	userMap := make(map[string]User)
	for _, nu := range config.Users {
		userMap[nu.Name] = nu.User
	}

	// Parse each context
	var parsedContexts []ParsedContext
	for _, nc := range config.Contexts {
		ctx := nc.Context
		cluster, ok := clusterMap[ctx.Cluster]
		if !ok {
			return nil, ValidationError{
				Field:   fmt.Sprintf("context.%s.cluster", nc.Name),
				Message: fmt.Sprintf("cluster '%s' not found", ctx.Cluster),
			}
		}

		user, ok := userMap[ctx.User]
		if !ok {
			return nil, ValidationError{
				Field:   fmt.Sprintf("context.%s.user", nc.Name),
				Message: fmt.Sprintf("user '%s' not found", ctx.User),
			}
		}

		// Validate required fields
		if cluster.Server == "" {
			return nil, ValidationError{
				Field:   fmt.Sprintf("cluster.%s.server", ctx.Cluster),
				Message: "server URL is required",
			}
		}

		parsed := ParsedContext{
			Name:                     nc.Name,
			ClusterName:              ctx.Cluster,
			Server:                   cluster.Server,
			CertificateAuthorityData: cluster.CertificateAuthorityData,
			UserName:                 ctx.User,
			Token:                    user.Token,
			ClientCertificateData:    user.ClientCertificateData,
			ClientKeyData:            user.ClientKeyData,
		}

		parsedContexts = append(parsedContexts, parsed)
	}

	if len(parsedContexts) == 0 {
		return nil, ValidationError{
			Field:   "contexts",
			Message: "no contexts found in kubeconfig",
		}
	}

	return parsedContexts, nil
}

// validateKubeconfig validates the basic structure of a kubeconfig
func validateKubeconfig(config *KubeconfigFile) error {
	if config.APIVersion == "" {
		return ValidationError{
			Field:   "apiVersion",
			Message: "apiVersion is required",
		}
	}

	if config.Kind != "Config" {
		return ValidationError{
			Field:   "kind",
			Message: "kind must be 'Config'",
		}
	}

	if len(config.Clusters) == 0 {
		return ValidationError{
			Field:   "clusters",
			Message: "at least one cluster is required",
		}
	}

	if len(config.Users) == 0 {
		return ValidationError{
			Field:   "users",
			Message: "at least one user is required",
		}
	}

	if len(config.Contexts) == 0 {
		return ValidationError{
			Field:   "contexts",
			Message: "at least one context is required",
		}
	}

	// Check for duplicate cluster names
	clusterNames := make(map[string]bool)
	for _, nc := range config.Clusters {
		if nc.Name == "" {
			return ValidationError{
				Field:   "clusters",
				Message: "cluster name cannot be empty",
			}
		}
		if clusterNames[nc.Name] {
			return ValidationError{
				Field:   "clusters",
				Message: fmt.Sprintf("duplicate cluster name: %s", nc.Name),
			}
		}
		clusterNames[nc.Name] = true
	}

	// Check for duplicate user names
	userNames := make(map[string]bool)
	for _, nu := range config.Users {
		if nu.Name == "" {
			return ValidationError{
				Field:   "users",
				Message: "user name cannot be empty",
			}
		}
		if userNames[nu.Name] {
			return ValidationError{
				Field:   "users",
				Message: fmt.Sprintf("duplicate user name: %s", nu.Name),
			}
		}
		userNames[nu.Name] = true
	}

	// Check for duplicate context names
	contextNames := make(map[string]bool)
	for _, nc := range config.Contexts {
		if nc.Name == "" {
			return ValidationError{
				Field:   "contexts",
				Message: "context name cannot be empty",
			}
		}
		if contextNames[nc.Name] {
			return ValidationError{
				Field:   "contexts",
				Message: fmt.Sprintf("duplicate context name: %s", nc.Name),
			}
		}
		contextNames[nc.Name] = true
	}

	return nil
}
