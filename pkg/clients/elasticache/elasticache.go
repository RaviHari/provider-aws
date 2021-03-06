/*
Copyright 2019 The Crossplane Authors.

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

package elasticache

import (
	"context"
	"reflect"
	"strconv"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/mitchellh/copystructure"
	"github.com/pkg/errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	"github.com/aws/aws-sdk-go-v2/service/elasticache/elasticacheiface"

	"github.com/crossplane/provider-aws/apis/cache/v1alpha1"
	cachev1alpha1 "github.com/crossplane/provider-aws/apis/cache/v1alpha1"
	"github.com/crossplane/provider-aws/apis/cache/v1beta1"
	clients "github.com/crossplane/provider-aws/pkg/clients"
)

const errCheckUpToDate = "unable to determine if external resource is up to date"

// A Client handles CRUD operations for ElastiCache resources. This interface is
// compatible with the upstream AWS redis client.
type Client elasticacheiface.ClientAPI

// NewClient returns a new ElastiCache client. Credentials must be passed as
// JSON encoded data.
func NewClient(ctx context.Context, credentials []byte, region string, auth clients.AuthMethod) (Client, error) {
	cfg, err := auth(ctx, credentials, clients.DefaultSection, region)
	if cfg == nil {
		return nil, err
	}
	return elasticache.New(*cfg), err
}

// TODO(negz): Determine whether we have to handle converting zero values to
// nil for the below types.

// NewCreateReplicationGroupInput returns ElastiCache replication group creation
// input suitable for use with the AWS API.
func NewCreateReplicationGroupInput(g v1beta1.ReplicationGroupParameters, id string, authToken *string) *elasticache.CreateReplicationGroupInput {
	c := &elasticache.CreateReplicationGroupInput{
		ReplicationGroupId:          &id,
		ReplicationGroupDescription: &g.ReplicationGroupDescription,
		Engine:                      &g.Engine,
		CacheNodeType:               &g.CacheNodeType,

		AtRestEncryptionEnabled:    g.AtRestEncryptionEnabled,
		AuthToken:                  authToken,
		AutomaticFailoverEnabled:   g.AutomaticFailoverEnabled,
		CacheParameterGroupName:    g.CacheParameterGroupName,
		CacheSecurityGroupNames:    g.CacheSecurityGroupNames,
		CacheSubnetGroupName:       g.CacheSubnetGroupName,
		EngineVersion:              g.EngineVersion,
		NotificationTopicArn:       g.NotificationTopicARN,
		NumCacheClusters:           clients.Int64Address(g.NumCacheClusters),
		NumNodeGroups:              clients.Int64Address(g.NumNodeGroups),
		Port:                       clients.Int64Address(g.Port),
		PreferredCacheClusterAZs:   g.PreferredCacheClusterAZs,
		PreferredMaintenanceWindow: g.PreferredMaintenanceWindow,
		PrimaryClusterId:           g.PrimaryClusterID,
		ReplicasPerNodeGroup:       clients.Int64Address(g.ReplicasPerNodeGroup),
		SecurityGroupIds:           g.SecurityGroupIDs,
		SnapshotArns:               g.SnapshotARNs,
		SnapshotName:               g.SnapshotName,
		SnapshotRetentionLimit:     clients.Int64Address(g.SnapshotRetentionLimit),
		SnapshotWindow:             g.SnapshotWindow,
		TransitEncryptionEnabled:   g.TransitEncryptionEnabled,
	}
	if len(g.Tags) != 0 {
		c.Tags = make([]elasticache.Tag, len(g.Tags))
		for i, tag := range g.Tags {
			c.Tags[i] = elasticache.Tag{
				Key:   clients.String(tag.Key),
				Value: clients.String(tag.Value),
			}
		}
	}
	if len(g.NodeGroupConfiguration) != 0 {
		c.NodeGroupConfiguration = make([]elasticache.NodeGroupConfiguration, len(g.NodeGroupConfiguration))
		for i, cfg := range g.NodeGroupConfiguration {
			c.NodeGroupConfiguration[i] = elasticache.NodeGroupConfiguration{
				PrimaryAvailabilityZone:  cfg.PrimaryAvailabilityZone,
				ReplicaAvailabilityZones: cfg.ReplicaAvailabilityZones,
				ReplicaCount:             clients.Int64Address(cfg.ReplicaCount),
				Slots:                    cfg.Slots,
			}
		}
	}
	return c
}

// NewModifyReplicationGroupInput returns ElastiCache replication group
// modification input suitable for use with the AWS API.
func NewModifyReplicationGroupInput(g v1beta1.ReplicationGroupParameters, id string) *elasticache.ModifyReplicationGroupInput {
	return &elasticache.ModifyReplicationGroupInput{
		ReplicationGroupId:          aws.String(id),
		ApplyImmediately:            aws.Bool(g.ApplyModificationsImmediately),
		AutomaticFailoverEnabled:    g.AutomaticFailoverEnabled,
		CacheNodeType:               aws.String(g.CacheNodeType),
		CacheParameterGroupName:     g.CacheParameterGroupName,
		CacheSecurityGroupNames:     g.CacheSecurityGroupNames,
		EngineVersion:               g.EngineVersion,
		NotificationTopicArn:        g.NotificationTopicARN,
		NotificationTopicStatus:     g.NotificationTopicStatus,
		PreferredMaintenanceWindow:  g.PreferredMaintenanceWindow,
		PrimaryClusterId:            g.PrimaryClusterID,
		ReplicationGroupDescription: aws.String(g.ReplicationGroupDescription),
		SecurityGroupIds:            g.SecurityGroupIDs,
		SnapshotRetentionLimit:      clients.Int64Address(g.SnapshotRetentionLimit),
		SnapshotWindow:              g.SnapshotWindow,
		SnapshottingClusterId:       g.SnapshottingClusterID,
	}
}

// NewDeleteReplicationGroupInput returns ElastiCache replication group deletion
// input suitable for use with the AWS API.
func NewDeleteReplicationGroupInput(id string) *elasticache.DeleteReplicationGroupInput {
	return &elasticache.DeleteReplicationGroupInput{ReplicationGroupId: &id}
}

// NewDescribeReplicationGroupsInput returns ElastiCache replication group describe
// input suitable for use with the AWS API.
func NewDescribeReplicationGroupsInput(id string) *elasticache.DescribeReplicationGroupsInput {
	return &elasticache.DescribeReplicationGroupsInput{ReplicationGroupId: &id}
}

// NewDescribeCacheClustersInput returns ElastiCache cache cluster describe
// input suitable for use with the AWS API.
func NewDescribeCacheClustersInput(clusterID string) *elasticache.DescribeCacheClustersInput {
	return &elasticache.DescribeCacheClustersInput{CacheClusterId: &clusterID}
}

// LateInitialize assigns the observed configurations and assigns them to the
// corresponding fields in ReplicationGroupParameters in order to let user
// know the defaults and make the changes as wished on that value.
func LateInitialize(s *v1beta1.ReplicationGroupParameters, rg elasticache.ReplicationGroup, cc elasticache.CacheCluster) {
	if s == nil {
		return
	}
	s.AtRestEncryptionEnabled = clients.LateInitializeBoolPtr(s.AtRestEncryptionEnabled, rg.AtRestEncryptionEnabled)
	s.AuthEnabled = clients.LateInitializeBoolPtr(s.AuthEnabled, rg.AuthTokenEnabled)
	s.AutomaticFailoverEnabled = clients.LateInitializeBoolPtr(s.AutomaticFailoverEnabled, automaticFailoverEnabled(rg.AutomaticFailover))
	s.SnapshotRetentionLimit = clients.LateInitializeIntPtr(s.SnapshotRetentionLimit, rg.SnapshotRetentionLimit)
	s.SnapshotWindow = clients.LateInitializeStringPtr(s.SnapshotWindow, rg.SnapshotWindow)
	s.SnapshottingClusterID = clients.LateInitializeStringPtr(s.SnapshottingClusterID, rg.SnapshottingClusterId)
	s.TransitEncryptionEnabled = clients.LateInitializeBoolPtr(s.TransitEncryptionEnabled, rg.TransitEncryptionEnabled)

	// NOTE(muvaf): ReplicationGroup managed N identical CacheCluster objects.
	// While configuration of those CacheClusters flow through ReplicationGroup API,
	// their statuses are fetched independently. Since we check for drifts against
	// the current state, late-init and up-to-date checks have to be made against
	// CacheClusters as well.
	s.EngineVersion = clients.LateInitializeStringPtr(s.EngineVersion, cc.EngineVersion)
	if cc.CacheParameterGroup != nil {
		s.CacheParameterGroupName = clients.LateInitializeStringPtr(s.CacheParameterGroupName, cc.CacheParameterGroup.CacheParameterGroupName)
	}
	if cc.NotificationConfiguration != nil {
		s.NotificationTopicARN = clients.LateInitializeStringPtr(s.NotificationTopicARN, cc.NotificationConfiguration.TopicArn)
		s.NotificationTopicStatus = clients.LateInitializeStringPtr(s.NotificationTopicStatus, cc.NotificationConfiguration.TopicStatus)
	}
	s.PreferredMaintenanceWindow = clients.LateInitializeStringPtr(s.PreferredMaintenanceWindow, cc.PreferredMaintenanceWindow)
	if len(s.SecurityGroupIDs) == 0 && len(cc.SecurityGroups) != 0 {
		s.SecurityGroupIDs = make([]string, len(cc.SecurityGroups))
		for i, val := range cc.SecurityGroups {
			s.SecurityGroupIDs[i] = aws.StringValue(val.SecurityGroupId)
		}
	}
	if len(s.CacheSecurityGroupNames) == 0 && len(cc.CacheSecurityGroups) != 0 {
		s.CacheSecurityGroupNames = make([]string, len(cc.CacheSecurityGroups))
		for i, val := range cc.CacheSecurityGroups {
			s.CacheSecurityGroupNames[i] = aws.StringValue(val.CacheSecurityGroupName)
		}
	}
}

// ReplicationGroupNeedsUpdate returns true if the supplied ReplicationGroup and
// the configuration of its member clusters differ from given desired state.
func ReplicationGroupNeedsUpdate(kube v1beta1.ReplicationGroupParameters, rg elasticache.ReplicationGroup, ccList []elasticache.CacheCluster) bool {
	switch {
	case !reflect.DeepEqual(kube.AutomaticFailoverEnabled, automaticFailoverEnabled(rg.AutomaticFailover)):
		return true
	case !reflect.DeepEqual(&kube.CacheNodeType, rg.CacheNodeType):
		return true
	case !reflect.DeepEqual(kube.SnapshotRetentionLimit, clients.IntAddress(rg.SnapshotRetentionLimit)):
		return true
	case !reflect.DeepEqual(kube.SnapshotWindow, rg.SnapshotWindow):
		return true
	}
	for _, cc := range ccList {
		if cacheClusterNeedsUpdate(kube, cc) {
			return true
		}
	}
	return false
}

func automaticFailoverEnabled(af elasticache.AutomaticFailoverStatus) *bool {
	if af == "" {
		return nil
	}
	r := af == elasticache.AutomaticFailoverStatusEnabled || af == elasticache.AutomaticFailoverStatusEnabling
	return &r
}

func cacheClusterNeedsUpdate(kube v1beta1.ReplicationGroupParameters, cc elasticache.CacheCluster) bool { // nolint:gocyclo
	// AWS will set and return a default version if we don't specify one.
	if !reflect.DeepEqual(kube.EngineVersion, cc.EngineVersion) {
		return true
	}
	if pg, name := cc.CacheParameterGroup, kube.CacheParameterGroupName; pg != nil && !reflect.DeepEqual(name, pg.CacheParameterGroupName) {
		return true
	}
	if cc.NotificationConfiguration != nil {
		if !reflect.DeepEqual(kube.NotificationTopicARN, cc.NotificationConfiguration.TopicArn) {
			return true
		}
		if !reflect.DeepEqual(cc.NotificationConfiguration.TopicStatus, kube.NotificationTopicStatus) {
			return true
		}
	} else if clients.StringValue(kube.NotificationTopicARN) != "" {
		return true
	}
	if !reflect.DeepEqual(kube.PreferredMaintenanceWindow, cc.PreferredMaintenanceWindow) {
		return true
	}
	return sgIDsNeedUpdate(kube.SecurityGroupIDs, cc.SecurityGroups) || sgNamesNeedUpdate(kube.CacheSecurityGroupNames, cc.CacheSecurityGroups)
}

func sgIDsNeedUpdate(kube []string, cc []elasticache.SecurityGroupMembership) bool {
	if len(kube) != len(cc) {
		return true
	}
	existingOnes := map[string]bool{}
	for _, sg := range cc {
		existingOnes[clients.StringValue(sg.SecurityGroupId)] = true
	}
	for _, desired := range kube {
		if !existingOnes[desired] {
			return true
		}
	}
	return false
}

func sgNamesNeedUpdate(kube []string, cc []elasticache.CacheSecurityGroupMembership) bool {
	if len(kube) != len(cc) {
		return true
	}
	existingOnes := map[string]bool{}
	for _, sg := range cc {
		existingOnes[clients.StringValue(sg.CacheSecurityGroupName)] = true
	}
	for _, desired := range kube {
		if !existingOnes[desired] {
			return true
		}
	}
	return false
}

// GenerateObservation produces a ReplicationGroupObservation object out of
// received elasticache.ReplicationGroup object.
func GenerateObservation(rg elasticache.ReplicationGroup) v1beta1.ReplicationGroupObservation {
	o := v1beta1.ReplicationGroupObservation{
		AutomaticFailover:     string(rg.AutomaticFailover),
		ClusterEnabled:        aws.BoolValue(rg.ClusterEnabled),
		ConfigurationEndpoint: newEndpoint(rg.ConfigurationEndpoint),
		MemberClusters:        rg.MemberClusters,
		Status:                clients.StringValue(rg.Status),
	}
	if len(rg.NodeGroups) != 0 {
		o.NodeGroups = make([]v1beta1.NodeGroup, len(rg.NodeGroups))
		for i, ng := range rg.NodeGroups {
			o.NodeGroups[i] = generateNodeGroup(ng)
		}
	}
	if rg.PendingModifiedValues != nil {
		o.PendingModifiedValues = generateReplicationGroupPendingModifiedValues(*rg.PendingModifiedValues)
	}
	return o
}

func generateNodeGroup(ng elasticache.NodeGroup) v1beta1.NodeGroup {
	r := v1beta1.NodeGroup{
		NodeGroupID: clients.StringValue(ng.NodeGroupId),
		Slots:       clients.StringValue(ng.Slots),
		Status:      clients.StringValue(ng.Status),
	}
	if len(ng.NodeGroupMembers) != 0 {
		r.NodeGroupMembers = make([]v1beta1.NodeGroupMember, len(ng.NodeGroupMembers))
		for i, m := range ng.NodeGroupMembers {
			r.NodeGroupMembers[i] = v1beta1.NodeGroupMember{
				CacheClusterID:            clients.StringValue(m.CacheClusterId),
				CacheNodeID:               clients.StringValue(m.CacheNodeId),
				CurrentRole:               clients.StringValue(m.CurrentRole),
				PreferredAvailabilityZone: clients.StringValue(m.PreferredAvailabilityZone),
			}
			if m.ReadEndpoint != nil {
				r.NodeGroupMembers[i].ReadEndpoint = v1beta1.Endpoint{
					Address: clients.StringValue(m.ReadEndpoint.Address),
					Port:    int(aws.Int64Value(m.ReadEndpoint.Port)),
				}
			}
		}
	}
	return r
}

func generateReplicationGroupPendingModifiedValues(in elasticache.ReplicationGroupPendingModifiedValues) v1beta1.ReplicationGroupPendingModifiedValues {
	r := v1beta1.ReplicationGroupPendingModifiedValues{
		AutomaticFailoverStatus: string(in.AutomaticFailoverStatus),
		PrimaryClusterID:        clients.StringValue(in.PrimaryClusterId),
	}
	if in.Resharding != nil && in.Resharding.SlotMigration != nil {
		r.Resharding = v1beta1.ReshardingStatus{
			SlotMigration: v1beta1.SlotMigration{
				ProgressPercentage: int(aws.Float64Value(in.Resharding.SlotMigration.ProgressPercentage)),
			},
		}
	}
	return r
}

func newEndpoint(e *elasticache.Endpoint) v1beta1.Endpoint {
	if e == nil {
		return v1beta1.Endpoint{}
	}

	return v1beta1.Endpoint{Address: clients.StringValue(e.Address), Port: int(aws.Int64Value(e.Port))}
}

// ConnectionEndpoint returns the connection endpoint for a Replication Group.
// https://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/Endpoints.html
func ConnectionEndpoint(rg elasticache.ReplicationGroup) managed.ConnectionDetails {
	// "Cluster enabled" Replication Groups have multiple node groups, and an
	// explicit configuration endpoint that should be used for read and write.
	if aws.BoolValue(rg.ClusterEnabled) &&
		rg.ConfigurationEndpoint != nil &&
		rg.ConfigurationEndpoint.Address != nil {
		return managed.ConnectionDetails{
			runtimev1alpha1.ResourceCredentialsSecretEndpointKey: []byte(aws.StringValue(rg.ConfigurationEndpoint.Address)),
			runtimev1alpha1.ResourceCredentialsSecretPortKey:     []byte(strconv.Itoa(int(aws.Int64Value(rg.ConfigurationEndpoint.Port)))),
		}
	}

	// "Cluster disabled" Replication Groups have a single node group, with a
	// primary endpoint that should be used for write. Any node's endpoint can
	// be used for read, but we support only a single endpoint so we return the
	// primary's.
	if len(rg.NodeGroups) > 0 &&
		rg.NodeGroups[0].PrimaryEndpoint != nil &&
		rg.NodeGroups[0].PrimaryEndpoint.Address != nil {
		return managed.ConnectionDetails{
			runtimev1alpha1.ResourceCredentialsSecretEndpointKey: []byte(aws.StringValue(rg.NodeGroups[0].PrimaryEndpoint.Address)),
			runtimev1alpha1.ResourceCredentialsSecretPortKey:     []byte(strconv.Itoa(int(aws.Int64Value(rg.NodeGroups[0].PrimaryEndpoint.Port)))),
		}
	}

	// If the AWS API docs are to be believed we should never get here.
	return nil
}

// IsNotFound returns true if the supplied error indicates a Replication Group
// was not found.
func IsNotFound(err error) bool {
	return isErrorCodeEqual(elasticache.ErrCodeReplicationGroupNotFoundFault, err)
}

// IsSubnetGroupNotFound returns true if the supplied error indicates a Cache Subnet Group
// was not found.
func IsSubnetGroupNotFound(err error) bool {
	return isErrorCodeEqual(elasticache.ErrCodeCacheSubnetGroupNotFoundFault, err)
}

// IsAlreadyExists returns true if the supplied error indicates a Replication Group
// already exists.
func IsAlreadyExists(err error) bool {
	return isErrorCodeEqual(elasticache.ErrCodeReplicationGroupAlreadyExistsFault, err)
}

func isErrorCodeEqual(errorCode string, err error) bool {
	ce, ok := err.(interface {
		Code() string
	})
	if !ok {
		return false
	}

	return ce.Code() == errorCode
}

// IsSubnetGroupUpToDate checks if CacheSubnetGroupParameters are in sync with provider values
func IsSubnetGroupUpToDate(p cachev1alpha1.CacheSubnetGroupParameters, sg elasticache.CacheSubnetGroup) bool {
	if p.Description != aws.StringValue(sg.CacheSubnetGroupDescription) {
		return false
	}

	if len(p.SubnetIDs) != len(sg.Subnets) {
		return false
	}

	exists := make(map[string]bool)
	for _, s := range sg.Subnets {
		exists[*s.SubnetIdentifier] = true
	}
	for _, id := range p.SubnetIDs {
		if !exists[id] {
			return false
		}
	}

	return true
}

// GenerateCreateCacheClusterInput returns Cache Cluster creation input
func GenerateCreateCacheClusterInput(p cachev1alpha1.CacheClusterParameters, id string) *elasticache.CreateCacheClusterInput {
	c := &elasticache.CreateCacheClusterInput{
		AZMode:                     elasticache.AZMode(aws.StringValue(p.AZMode)),
		AuthToken:                  p.AZMode,
		CacheClusterId:             aws.String(id),
		CacheNodeType:              aws.String(p.CacheNodeType),
		CacheParameterGroupName:    p.CacheParameterGroupName,
		CacheSubnetGroupName:       p.CacheSubnetGroupName,
		CacheSecurityGroupNames:    p.CacheSecurityGroupNames,
		Engine:                     p.Engine,
		EngineVersion:              p.EngineVersion,
		NotificationTopicArn:       p.NotificationTopicARN,
		NumCacheNodes:              aws.Int64(p.NumCacheNodes),
		Port:                       p.Port,
		PreferredAvailabilityZone:  p.PreferredAvailabilityZone,
		PreferredAvailabilityZones: p.PreferredAvailabilityZones,
		PreferredMaintenanceWindow: p.PreferredMaintenanceWindow,
		ReplicationGroupId:         p.ReplicationGroupID,
		SecurityGroupIds:           p.SecurityGroupIDs,
		SnapshotArns:               p.SnapshotARNs,
		SnapshotName:               p.SnapshotName,
		SnapshotRetentionLimit:     p.SnapshotRetentionLimit,
		SnapshotWindow:             p.SnapshotWindow,
	}

	if len(p.Tags) != 0 {
		c.Tags = make([]elasticache.Tag, len(p.Tags))
		for i, tag := range p.Tags {
			c.Tags[i] = elasticache.Tag{
				Key:   clients.String(tag.Key),
				Value: tag.Value,
			}
		}
	}

	return c
}

// GenerateModifyCacheClusterInput returns ElastiCache Cache Cluster
// modification input suitable for use with the AWS API.
func GenerateModifyCacheClusterInput(p cachev1alpha1.CacheClusterParameters, id string) *elasticache.ModifyCacheClusterInput {
	return &elasticache.ModifyCacheClusterInput{
		CacheClusterId:             aws.String(id),
		AZMode:                     elasticache.AZMode(aws.StringValue(p.AZMode)),
		ApplyImmediately:           p.ApplyImmediately,
		AuthToken:                  p.AuthToken,
		AuthTokenUpdateStrategy:    elasticache.AuthTokenUpdateStrategyType(clients.StringValue(p.AuthTokenUpdateStrategy)),
		CacheNodeIdsToRemove:       p.CacheNodeIDsToRemove,
		CacheNodeType:              aws.String(p.CacheNodeType),
		CacheParameterGroupName:    p.CacheParameterGroupName,
		CacheSecurityGroupNames:    p.CacheSecurityGroupNames,
		EngineVersion:              p.EngineVersion,
		NewAvailabilityZones:       p.PreferredAvailabilityZones,
		NotificationTopicArn:       p.NotificationTopicARN,
		NumCacheNodes:              aws.Int64(p.NumCacheNodes),
		PreferredMaintenanceWindow: p.PreferredMaintenanceWindow,
		SecurityGroupIds:           p.SecurityGroupIDs,
		SnapshotRetentionLimit:     p.SnapshotRetentionLimit,
		SnapshotWindow:             p.SnapshotWindow,
	}
}

// GenerateClusterObservation produces a CacheClusterObservation object out of
// received elasticache.CacheCluster object.
func GenerateClusterObservation(c elasticache.CacheCluster) cachev1alpha1.CacheClusterObservation {
	o := cachev1alpha1.CacheClusterObservation{
		AtRestEncryptionEnabled:   aws.BoolValue(c.AtRestEncryptionEnabled),
		AuthTokenEnabled:          aws.BoolValue(c.AtRestEncryptionEnabled),
		CacheClusterStatus:        aws.StringValue(c.CacheClusterStatus),
		ClientDownloadLandingPage: aws.StringValue(c.ClientDownloadLandingPage),
	}

	if len(c.CacheNodes) > 0 {
		cacheNodes := make([]v1alpha1.CacheNode, len(c.CacheNodes))
		for i, v := range c.CacheNodes {
			cacheNodes[i] = v1alpha1.CacheNode{
				CacheNodeID:              aws.StringValue(v.CacheNodeId),
				CacheNodeStatus:          aws.StringValue(v.CacheNodeStatus),
				CustomerAvailabilityZone: aws.StringValue(v.CustomerAvailabilityZone),
				ParameterGroupStatus:     aws.StringValue(v.ParameterGroupStatus),
				SourceCacheNodeID:        v.SourceCacheNodeId,
			}
			if v.Endpoint != nil {
				cacheNodes[i].Endpoint = &v1alpha1.Endpoint{
					Address: aws.StringValue(v.Endpoint.Address),
					Port:    int(aws.Int64Value(v.Endpoint.Port)),
				}
			}
		}
		o.CacheNodes = cacheNodes
	}
	return o
}

// IsClusterNotFound returns true if the supplied error indicates a Cache Cluster
// already exists.
func IsClusterNotFound(err error) bool {
	return isErrorCodeEqual(elasticache.ErrCodeCacheClusterNotFoundFault, err)
}

// LateInitializeCluster assigns the observed configurations and assigns them to the
// corresponding fields in CacheClusterParameters in order to let user
// know the defaults and make the changes as wished on that value.
func LateInitializeCluster(p *cachev1alpha1.CacheClusterParameters, c elasticache.CacheCluster) {
	p.SnapshotRetentionLimit = clients.LateInitializeInt64Ptr(p.SnapshotRetentionLimit, c.SnapshotRetentionLimit)
	p.SnapshotWindow = clients.LateInitializeStringPtr(p.SnapshotWindow, c.SnapshotWindow)
	p.CacheSubnetGroupName = clients.LateInitializeStringPtr(p.CacheSubnetGroupName, c.CacheSubnetGroupName)
	p.EngineVersion = clients.LateInitializeStringPtr(p.EngineVersion, c.EngineVersion)
	p.PreferredAvailabilityZone = clients.LateInitializeStringPtr(p.PreferredAvailabilityZone, c.PreferredAvailabilityZone)
	p.PreferredMaintenanceWindow = clients.LateInitializeStringPtr(p.PreferredMaintenanceWindow, c.PreferredMaintenanceWindow)
	p.ReplicationGroupID = clients.LateInitializeStringPtr(p.ReplicationGroupID, c.ReplicationGroupId)

	if c.NotificationConfiguration != nil {
		p.NotificationTopicARN = clients.LateInitializeStringPtr(p.NotificationTopicARN, c.NotificationConfiguration.TopicArn)
	}
	if c.CacheParameterGroup != nil {
		p.CacheParameterGroupName = clients.LateInitializeStringPtr(p.CacheParameterGroupName, c.CacheParameterGroup.CacheParameterGroupName)
	}
}

// GenerateCluster modifies elasticache.CacheCluster with values from cachev1alpha1.CacheClusterParameters
func GenerateCluster(name string, p cachev1alpha1.CacheClusterParameters, c *elasticache.CacheCluster) {
	c.CacheClusterId = aws.String(name)
	c.CacheNodeType = aws.String(p.CacheNodeType)
	c.EngineVersion = p.EngineVersion
	c.NumCacheNodes = aws.Int64(p.NumCacheNodes)
	c.PreferredMaintenanceWindow = p.PreferredMaintenanceWindow
	c.SnapshotRetentionLimit = p.SnapshotRetentionLimit
	c.SnapshotWindow = p.SnapshotWindow

	if len(p.SecurityGroupIDs) > 0 {
		sg := make([]elasticache.SecurityGroupMembership, len(p.SecurityGroupIDs))
		for i, v := range p.SecurityGroupIDs {
			sg[i] = elasticache.SecurityGroupMembership{
				SecurityGroupId: aws.String(v),
				Status:          aws.String("active"),
			}
		}
		c.SecurityGroups = sg
	}

	if c.CacheParameterGroup != nil {
		c.CacheParameterGroup.CacheParameterGroupName = p.CacheParameterGroupName
	}

	if c.NotificationConfiguration != nil {
		c.NotificationConfiguration.TopicArn = p.NotificationTopicARN
	}
}

// IsClusterUpToDate checks whether current state is up-to-date compared to the given
// set of parameters.
func IsClusterUpToDate(name string, in *cachev1alpha1.CacheClusterParameters, observed *elasticache.CacheCluster) (bool, error) {
	generated, err := copystructure.Copy(observed)
	if err != nil {
		return true, errors.Wrap(err, errCheckUpToDate)
	}
	desired, ok := generated.(*elasticache.CacheCluster)
	if !ok {
		return true, errors.New(errCheckUpToDate)
	}
	GenerateCluster(name, *in, desired)

	return cmp.Equal(desired, observed, cmpopts.EquateEmpty()), nil
}
