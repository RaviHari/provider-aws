/*
Copyright 2020 The Crossplane Authors.

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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
)

// ClusterStatusType is the status of an EKS cluster.
type ClusterStatusType string

// Cluster statuses.
const (
	ClusterStatusCreating ClusterStatusType = "CREATING"
	ClusterStatusActive   ClusterStatusType = "ACTIVE"
	ClusterStatusDeleting ClusterStatusType = "DELETING"
	ClusterStatusFailed   ClusterStatusType = "FAILED"
	ClusterStatusUpdating ClusterStatusType = "UPDATING"
)

// LogType is a type of logging.
type LogType string

// Log types.
const (
	LogTypeAPI               LogType = "api"
	LogTypeAudit             LogType = "audit"
	LogTypeAuthenticator     LogType = "authenticator"
	LogTypeControllerManager LogType = "controllerManager"
	LogTypeScheduler         LogType = "scheduler"
)

// ClusterParameters define the desired state of an AWS Elastic Kubernetes
// Service cluster.
type ClusterParameters struct {
	// The encryption configuration for the cluster.
	// +immutable
	// +optional
	EncryptionConfig []EncryptionConfig `json:"encryptionConfig,omitempty"`

	// Enable or disable exporting the Kubernetes control plane logs for your cluster
	// to CloudWatch Logs. By default, cluster control plane logs aren't exported
	// to CloudWatch Logs. For more information, see Amazon EKS Cluster Control
	// Plane Logs (https://docs.aws.amazon.com/eks/latest/userguide/control-plane-logs.html)
	// in the Amazon EKS User Guide .
	//
	// CloudWatch Logs ingestion, archive storage, and data scanning rates apply
	// to exported control plane logs. For more information, see Amazon CloudWatch
	// Pricing (http://aws.amazon.com/cloudwatch/pricing/).
	// +optional
	Logging *Logging `json:"logging,omitempty"`

	// The VPC configuration used by the cluster control plane. Amazon EKS VPC resources
	// have specific requirements to work properly with Kubernetes. For more information,
	// see Cluster VPC Considerations (https://docs.aws.amazon.com/eks/latest/userguide/network_reqs.html)
	// and Cluster Security Group Considerations (https://docs.aws.amazon.com/eks/latest/userguide/sec-group-reqs.html)
	// in the Amazon EKS User Guide. You must specify at least two subnets. You
	// can specify up to five security groups, but we recommend that you use a dedicated
	// security group for your cluster control plane.
	//
	// ResourcesVpcConfig is a required field
	ResourcesVpcConfig VpcConfigRequest `json:"resourcesVpcConfig"`

	// The Amazon Resource Name (ARN) of the IAM role that provides permissions
	// for Amazon EKS to make calls to other AWS API operations on your behalf.
	// For more information, see Amazon EKS Service IAM Role (https://docs.aws.amazon.com/eks/latest/userguide/service_IAM_role.html)
	// in the Amazon EKS User Guide .
	//
	// RoleArn is a required field
	// +immutable
	RoleArn string `json:"roleArn,omitempty"`

	// RoleArnRef is a reference to an IAMRole used to set
	// the RoleArn.
	// +immutable
	// +optional
	RoleArnRef *runtimev1alpha1.Reference `json:"roleArnRef,omitempty"`

	// RoleArnSelector selects references to IAMRole used
	// to set the RoleArn.
	// +optional
	RoleArnSelector *runtimev1alpha1.Selector `json:"roleArnSelector,omitempty"`

	// The metadata to apply to the cluster to assist with categorization and organization.
	// Each tag consists of a key and an optional value, both of which you define.
	// +optional
	Tags map[string]string `json:"tags,omitempty"`

	// The desired Kubernetes version for your cluster. If you don't specify a value
	// here, the latest version available in Amazon EKS is used.
	// Example: 1.15
	// +optional
	Version *string `json:"version,omitempty"`
}

// EncryptionConfig is the encryption configuration for a cluster.
type EncryptionConfig struct {

	// AWS Key Management Service (AWS KMS) customer master key (CMK). Either the
	// ARN or the alias can be used.
	Provider Provider `json:"provider"`

	// Specifies the resources to be encrypted. The only supported value is "secrets".
	Resources []string `json:"resources"`
}

// Provider is an encryption provider.
type Provider struct {

	// Amazon Resource Name (ARN) or alias of the customer master key (CMK). The
	// CMK must be symmetric, created in the same region as the cluster, and if
	// the CMK was created in a different account, the user must have access to
	// the CMK. For more information, see Allowing Users in Other Accounts to Use
	// a CMK (https://docs.aws.amazon.com/kms/latest/developerguide/key-policy-modifying-external-accounts.html)
	// in the AWS Key Management Service Developer Guide.
	KeyArn string `json:"keyArn"`
}

// Logging in the logging configuration for a cluster.
type Logging struct {
	// The cluster control plane logging configuration for your cluster.
	ClusterLogging []LogSetup `json:"clusterLogging"`
}

// LogSetup specifies the logging types that are enabled.
type LogSetup struct {
	// If a log type is enabled, that log type exports its control plane logs to
	// CloudWatch Logs. If a log type isn't enabled, that log type doesn't export
	// its control plane logs. Each individual log type can be enabled or disabled
	// independently.
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// The available cluster control plane log types.
	Types []LogType `json:"types,omitempty"`
}

// VpcConfigRequest specifies the VPC configuration for a cluster.
type VpcConfigRequest struct {
	// Set this value to true to enable private access for your cluster's Kubernetes
	// API server endpoint. If you enable private access, Kubernetes API requests
	// from within your cluster's VPC use the private VPC endpoint. The default
	// value for this parameter is false, which disables private access for your
	// Kubernetes API server. For more information, see Amazon EKS Cluster Endpoint
	// Access Control (https://docs.aws.amazon.com/eks/latest/userguide/cluster-endpoint.html)
	// in the Amazon EKS User Guide.
	// +optional
	EndpointPrivateAccess *bool `json:"endpointPrivateAccess,omitempty"`

	// Set this value to false to disable public access for your cluster's Kubernetes
	// API server endpoint. If you disable public access, your cluster's Kubernetes
	// API server can receive only requests from within the cluster VPC. The default
	// value for this parameter is true, which enables public access for your Kubernetes
	// API server. For more information, see Amazon EKS Cluster Endpoint Access
	// Control (https://docs.aws.amazon.com/eks/latest/userguide/cluster-endpoint.html)
	// in the Amazon EKS User Guide.
	// +optional
	EndpointPublicAccess *bool `json:"endpointPublicAccess,omitempty"`

	// The CIDR blocks that are allowed access to your cluster's public Kubernetes
	// API server endpoint. Communication to the endpoint from addresses outside
	// of the CIDR blocks that you specify is denied. The default value is 0.0.0.0/0.
	// If you've disabled private endpoint access and you have worker nodes or AWS
	// Fargate pods in the cluster, then ensure that you specify the necessary CIDR
	// blocks. For more information, see Amazon EKS Cluster Endpoint Access Control
	// (https://docs.aws.amazon.com/eks/latest/userguide/cluster-endpoint.html)
	// in the Amazon EKS User Guide.
	// +optional
	PublicAccessCidrs []string `json:"publicAccessCidrs,omitempty"`

	// Specify one or more security groups for the cross-account elastic network
	// interfaces that Amazon EKS creates to use to allow communication between
	// your worker nodes and the Kubernetes control plane. If you don't specify
	// a security group, the default security group for your VPC is used.
	// +optional
	SecurityGroupIDs []string `json:"securityGroupIds,omitempty"`

	// SecurityGroupIDRefs are references to SecurityGroups used to set
	// the SecurityGroupIDs.
	// +optional
	SecurityGroupIDRefs []runtimev1alpha1.Reference `json:"securityGroupIdRefs,omitempty"`

	// SecurityGroupIDSelector selects references to SecurityGroups used
	// to set the SecurityGroupIDs.
	// +optional
	SecurityGroupIDSelector *runtimev1alpha1.Selector `json:"securityGroupIdSelector,omitempty"`

	// Specify subnets for your Amazon EKS worker nodes. Amazon EKS creates cross-account
	// elastic network interfaces in these subnets to allow communication between
	// your worker nodes and the Kubernetes control plane.
	// +optional
	SubnetIDs []string `json:"subnetIds,omitempty"`

	// SubnetIDRefs are references to Subnets used to set
	// the SubnetIDs.
	// +optional
	SubnetIDRefs []runtimev1alpha1.Reference `json:"subnetIdRefs,omitempty"`

	// SubnetIDSelector selects references to Subnets used
	// to set the SubnetIDs.
	// +optional
	SubnetIDSelector *runtimev1alpha1.Selector `json:"subnetIdSelector,omitempty"`
}

// ClusterObservation is the observed state of a cluster.
type ClusterObservation struct {
	// The Amazon Resource Name (ARN) of the cluster.
	Arn string `json:"arn,omitempty"`

	// The Unix epoch timestamp in seconds for when the cluster was created.
	CreatedAt *metav1.Time `json:"createdAt,omitempty"`

	// The endpoint for your Kubernetes API server.
	Endpoint string `json:"endpoint,omitempty"`

	// The identity provider information for the cluster.
	Identity Identity `json:"identity,omitempty"`

	// The platform version of your Amazon EKS cluster. For more information, see
	// Platform Versions (https://docs.aws.amazon.com/eks/latest/userguide/platform-versions.html)
	// in the Amazon EKS User Guide .
	PlatformVersion string `json:"platformVersion,omitempty"`

	// The VPC configuration used by the cluster control plane. Amazon EKS VPC resources
	// have specific requirements to work properly with Kubernetes. For more information,
	// see Cluster VPC Considerations (https://docs.aws.amazon.com/eks/latest/userguide/network_reqs.html)
	// and Cluster Security Group Considerations (https://docs.aws.amazon.com/eks/latest/userguide/sec-group-reqs.html)
	// in the Amazon EKS User Guide.
	ResourcesVpcConfig VpcConfigResponse `json:"resourcesVpcConfig,omitempty"`

	// The current status of the cluster.
	Status ClusterStatusType `json:"status,omitempty"`
}

// Identity is the identity information for a cluster.
type Identity struct {

	// The OpenID Connect (https://openid.net/connect/) identity provider information
	// for the cluster.
	OIDC OIDC `json:"oidc,omitempty"`
}

// OIDC is the OpenID Connect issuer URL.
type OIDC struct {
	// The issuer URL for the OpenID Connect identity provider.
	Issuer string `json:"issuer,omitempty"`
}

// VpcConfigResponse is the observed VPC configuration for a cluster.
type VpcConfigResponse struct {
	// The cluster security group that was created by Amazon EKS for the cluster.
	// Managed node groups use this security group for control-plane-to-data-plane
	// communication.
	ClusterSecurityGroupID string `json:"clusterSecurityGroupId,omitempty"`

	// The VPC associated with your cluster.
	VpcID string `json:"vpcId,omitempty"`
}

// A ClusterSpec defines the desired state of an EKS Cluster.
type ClusterSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	ForProvider                  ClusterParameters `json:"forProvider"`
}

// A ClusterStatus represents the observed state of an EKS Cluster.
type ClusterStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
	AtProvider                     ClusterObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Cluster is a managed resource that represents an AWS Elastic Kubernetes
// Service cluster.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="RECLAIM-POLICY",type="string",JSONPath=".spec.reclaimPolicy"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterSpec   `json:"spec"`
	Status ClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ClusterList contains a list of Cluster items
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cluster `json:"items"`
}
