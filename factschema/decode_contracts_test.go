// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package factschema

import (
	awsv1 "github.com/eshu-hq/eshu-sdk/factschema/aws/v1"
	azurev1 "github.com/eshu-hq/eshu-sdk/factschema/azure/v1"
	cicdrunv1 "github.com/eshu-hq/eshu-sdk/factschema/cicdrun/v1"
	codedataflowv1 "github.com/eshu-hq/eshu-sdk/factschema/codedataflow/v1"
	codegraphv1 "github.com/eshu-hq/eshu-sdk/factschema/codegraph/v1"
	codeownersv1 "github.com/eshu-hq/eshu-sdk/factschema/codeowners/v1"
	documentationv1 "github.com/eshu-hq/eshu-sdk/factschema/documentation/v1"
	gcpv1 "github.com/eshu-hq/eshu-sdk/factschema/gcp/v1"
	iamv1 "github.com/eshu-hq/eshu-sdk/factschema/iam/v1"
	incidentv1 "github.com/eshu-hq/eshu-sdk/factschema/incident/v1"
	kuberneteslivev1 "github.com/eshu-hq/eshu-sdk/factschema/kuberneteslive/v1"
	observabilityv1 "github.com/eshu-hq/eshu-sdk/factschema/observability/v1"
	ociregistryv1 "github.com/eshu-hq/eshu-sdk/factschema/ociregistry/v1"
	packageregistryv1 "github.com/eshu-hq/eshu-sdk/factschema/packageregistry/v1"
	reducerderivedv1 "github.com/eshu-hq/eshu-sdk/factschema/reducerderived/v1"
	refOverlayV1 "github.com/eshu-hq/eshu-sdk/factschema/ref_overlay/v1"
	sbomv1 "github.com/eshu-hq/eshu-sdk/factschema/sbom/v1"
	scannerworkerv1 "github.com/eshu-hq/eshu-sdk/factschema/scannerworker/v1"
	secretsiamv1 "github.com/eshu-hq/eshu-sdk/factschema/secretsiam/v1"
	securityalertv1 "github.com/eshu-hq/eshu-sdk/factschema/securityalert/v1"
	semanticv1 "github.com/eshu-hq/eshu-sdk/factschema/semantic/v1"
	servicecatalogv1 "github.com/eshu-hq/eshu-sdk/factschema/servicecatalog/v1"
	submodulev1 "github.com/eshu-hq/eshu-sdk/factschema/submodule/v1"
	tfstatev1 "github.com/eshu-hq/eshu-sdk/factschema/terraformstate/v1"
	vulnerabilityv1 "github.com/eshu-hq/eshu-sdk/factschema/vulnerability/v1"
	vulnerabilitysuppressionv1 "github.com/eshu-hq/eshu-sdk/factschema/vulnerabilitysuppression/v1"
	workitemv1 "github.com/eshu-hq/eshu-sdk/factschema/workitem/v1"
	"reflect"
)

var payloadContracts = []struct {
	// factKind is the fact kind identifier, used only for test messages.
	factKind string
	// schemaFile is the schema's filename under schema/.
	schemaFile string
	// typ is the payload struct type whose reflectively derived key set must
	// match the generated schema.
	typ reflect.Type
}{
	{FactKindAWSResource, "aws_resource.v1.schema.json", reflect.TypeOf(awsv1.Resource{})},
	{FactKindAWSRelationship, "aws_relationship.v1.schema.json", reflect.TypeOf(awsv1.Relationship{})},
	{FactKindAWSDNSRecord, "aws_dns_record.v1.schema.json", reflect.TypeOf(awsv1.DNSRecord{})},
	{FactKindAWSImageReference, "aws_image_reference.v1.schema.json", reflect.TypeOf(awsv1.ImageReference{})},
	{FactKindAWSSecurityGroupRule, "aws_security_group_rule.v1.schema.json", reflect.TypeOf(awsv1.SecurityGroupRule{})},
	{FactKindAWSWarning, "aws_warning.v1.schema.json", reflect.TypeOf(awsv1.Warning{})},
	{FactKindEC2InstancePosture, "ec2_instance_posture.v1.schema.json", reflect.TypeOf(awsv1.EC2InstancePosture{})},
	{FactKindRDSInstancePosture, "rds_instance_posture.v1.schema.json", reflect.TypeOf(awsv1.RDSInstancePosture{})},
	{FactKindS3BucketPosture, "s3_bucket_posture.v1.schema.json", reflect.TypeOf(awsv1.S3BucketPosture{})},
	{FactKindS3ExternalPrincipalGrant, "s3_external_principal_grant.v1.schema.json", reflect.TypeOf(awsv1.S3ExternalPrincipalGrant{})},
	{FactKindAWSIAMPermission, "aws_iam_permission.v1.schema.json", reflect.TypeOf(iamv1.Permission{})},
	{FactKindAWSResourcePolicyPermission, "aws_resource_policy_permission.v1.schema.json", reflect.TypeOf(iamv1.ResourcePolicyPermission{})},
	{FactKindAWSIAMPrincipal, "aws_iam_principal.v1.schema.json", reflect.TypeOf(iamv1.Principal{})},
	{FactKindAWSIAMTrustPolicy, "aws_iam_trust_policy.v1.schema.json", reflect.TypeOf(secretsiamv1.AWSIAMTrustPolicy{})},
	{FactKindAWSIAMPermissionPolicy, "aws_iam_permission_policy.v1.schema.json", reflect.TypeOf(secretsiamv1.AWSIAMPermissionPolicy{})},
	{FactKindAWSIAMPolicyAttachment, "aws_iam_policy_attachment.v1.schema.json", reflect.TypeOf(secretsiamv1.AWSIAMPolicyAttachment{})},
	{FactKindAWSIAMPermissionBoundary, "aws_iam_permission_boundary.v1.schema.json", reflect.TypeOf(secretsiamv1.AWSIAMPermissionBoundary{})},
	{FactKindAWSIAMInstanceProfile, "aws_iam_instance_profile.v1.schema.json", reflect.TypeOf(secretsiamv1.AWSIAMInstanceProfile{})},
	{FactKindAWSIAMAccessAnalyzerFinding, "aws_iam_access_analyzer_finding.v1.schema.json", reflect.TypeOf(secretsiamv1.AWSIAMAccessAnalyzerFinding{})},
	{FactKindIncidentRecord, "incident.record.v1.schema.json", reflect.TypeOf(incidentv1.IncidentRecord{})},
	{FactKindIncidentLifecycleEvent, "incident.lifecycle_event.v1.schema.json", reflect.TypeOf(incidentv1.LifecycleEvent{})},
	{FactKindChangeRecord, "change.record.v1.schema.json", reflect.TypeOf(incidentv1.ChangeRecord{})},
	{FactKindIncidentRoutingAppliedPagerDutyResource, "incident_routing.applied_pagerduty_resource.v1.schema.json", reflect.TypeOf(incidentv1.AppliedPagerDutyResource{})},
	{FactKindIncidentRoutingAppliedAlertRoute, "incident_routing.applied_alert_route.v1.schema.json", reflect.TypeOf(incidentv1.AppliedAlertRoute{})},
	{FactKindIncidentRoutingObservedPagerDutyService, "incident_routing.observed_pagerduty_service.v1.schema.json", reflect.TypeOf(incidentv1.ObservedPagerDutyService{})},
	{FactKindIncidentRoutingObservedPagerDutyIntegration, "incident_routing.observed_pagerduty_integration.v1.schema.json", reflect.TypeOf(incidentv1.ObservedPagerDutyIntegration{})},
	{FactKindIncidentRoutingCoverageWarning, "incident_routing.coverage_warning.v1.schema.json", reflect.TypeOf(incidentv1.CoverageWarning{})},
	{FactKindGCPCloudResource, "gcp_cloud_resource.v1.schema.json", reflect.TypeOf(gcpv1.Resource{})},
	{FactKindGCPCloudRelationship, "gcp_cloud_relationship.v1.schema.json", reflect.TypeOf(gcpv1.Relationship{})},
	{FactKindGCPCollectionWarning, "gcp_collection_warning.v1.schema.json", reflect.TypeOf(gcpv1.CollectionWarning{})},
	{FactKindGCPDNSRecord, "gcp_dns_record.v1.schema.json", reflect.TypeOf(gcpv1.DNSRecord{})},
	{FactKindGCPIAMPolicyObservation, "gcp_iam_policy_observation.v1.schema.json", reflect.TypeOf(gcpv1.IAMPolicyObservation{})},
	{FactKindGCPTagObservation, "gcp_tag_observation.v1.schema.json", reflect.TypeOf(gcpv1.TagObservation{})},
	{FactKindGCPImageReference, "gcp_image_reference.v1.schema.json", reflect.TypeOf(gcpv1.ImageReference{})},
	{FactKindGCPIAMPrincipal, "gcp_iam_principal.v1.schema.json", reflect.TypeOf(secretsiamv1.GCPIAMPrincipal{})},
	{FactKindGCPIAMTrustPolicy, "gcp_iam_trust_policy.v1.schema.json", reflect.TypeOf(secretsiamv1.GCPIAMTrustPolicy{})},
	{FactKindGCPIAMPermissionPolicy, "gcp_iam_permission_policy.v1.schema.json", reflect.TypeOf(secretsiamv1.GCPIAMPermissionPolicy{})},
	{FactKindAzureCloudResource, "azure_cloud_resource.v1.schema.json", reflect.TypeOf(azurev1.CloudResource{})},
	{FactKindAzureCloudRelationship, "azure_cloud_relationship.v1.schema.json", reflect.TypeOf(azurev1.CloudRelationship{})},
	{FactKindAzureDNSRecord, "azure_dns_record.v1.schema.json", reflect.TypeOf(azurev1.DNSRecord{})},
	{FactKindAzureCollectionWarning, "azure_collection_warning.v1.schema.json", reflect.TypeOf(azurev1.CollectionWarning{})},
	{FactKindAzureTagObservation, "azure_tag_observation.v1.schema.json", reflect.TypeOf(azurev1.TagObservation{})},
	{FactKindAzureIdentityObservation, "azure_identity_observation.v1.schema.json", reflect.TypeOf(azurev1.IdentityObservation{})},
	{FactKindAzureResourceChange, "azure_resource_change.v1.schema.json", reflect.TypeOf(azurev1.ResourceChange{})},
	{FactKindAzureImageReference, "azure_image_reference.v1.schema.json", reflect.TypeOf(azurev1.ImageReference{})},
	{FactKindKubernetesLivePodTemplate, "kubernetes_live.pod_template.v1.schema.json", reflect.TypeOf(kuberneteslivev1.PodTemplate{})},
	{FactKindKubernetesLiveRelationship, "kubernetes_live.relationship.v1.schema.json", reflect.TypeOf(kuberneteslivev1.Relationship{})},
	{FactKindKubernetesLiveWarning, "kubernetes_live.warning.v1.schema.json", reflect.TypeOf(kuberneteslivev1.Warning{})},
	{FactKindKubernetesLiveNamespace, "kubernetes_live.namespace.v1.schema.json", reflect.TypeOf(kuberneteslivev1.Namespace{})},
	{FactKindOCIRegistryRepository, "oci_registry.repository.v1.schema.json", reflect.TypeOf(ociregistryv1.Repository{})},
	{FactKindOCIImageManifest, "oci_registry.image_manifest.v1.schema.json", reflect.TypeOf(ociregistryv1.ImageManifest{})},
	{FactKindOCIImageIndex, "oci_registry.image_index.v1.schema.json", reflect.TypeOf(ociregistryv1.ImageIndex{})},
	{FactKindOCIImageDescriptor, "oci_registry.image_descriptor.v1.schema.json", reflect.TypeOf(ociregistryv1.ImageDescriptor{})},
	{FactKindOCIImageTagObservation, "oci_registry.image_tag_observation.v1.schema.json", reflect.TypeOf(ociregistryv1.TagObservation{})},
	{FactKindOCIImageReferrer, "oci_registry.image_referrer.v1.schema.json", reflect.TypeOf(ociregistryv1.ImageReferrer{})},
	{FactKindOCIRegistryWarning, "oci_registry.warning.v1.schema.json", reflect.TypeOf(ociregistryv1.Warning{})},
	{FactKindTerraformStateSnapshot, "terraform_state_snapshot.v1.schema.json", reflect.TypeOf(tfstatev1.Snapshot{})},
	{FactKindTerraformStateResource, "terraform_state_resource.v1.schema.json", reflect.TypeOf(tfstatev1.Resource{})},
	{FactKindTerraformStateModule, "terraform_state_module.v1.schema.json", reflect.TypeOf(tfstatev1.Module{})},
	{FactKindTerraformStateOutput, "terraform_state_output.v1.schema.json", reflect.TypeOf(tfstatev1.Output{})},
	{FactKindTerraformStateTagObservation, "terraform_state_tag_observation.v1.schema.json", reflect.TypeOf(tfstatev1.TagObservation{})},
	{FactKindTerraformStateCandidate, "terraform_state_candidate.v1.schema.json", reflect.TypeOf(tfstatev1.Candidate{})},
	{FactKindTerraformStateProviderBinding, "terraform_state_provider_binding.v1.schema.json", reflect.TypeOf(tfstatev1.ProviderBinding{})},
	{FactKindTerraformStateWarning, "terraform_state_warning.v1.schema.json", reflect.TypeOf(tfstatev1.Warning{})},
	{FactKindPackageRegistryPackage, "package_registry.package.v1.schema.json", reflect.TypeOf(packageregistryv1.Package{})},
	{FactKindPackageRegistryPackageVersion, "package_registry.package_version.v1.schema.json", reflect.TypeOf(packageregistryv1.PackageVersion{})},
	{FactKindPackageRegistryPackageDependency, "package_registry.package_dependency.v1.schema.json", reflect.TypeOf(packageregistryv1.PackageDependency{})},
	{FactKindPackageRegistrySourceHint, "package_registry.source_hint.v1.schema.json", reflect.TypeOf(packageregistryv1.SourceHint{})},
	{FactKindPackageRegistryPackageArtifact, "package_registry.package_artifact.v1.schema.json", reflect.TypeOf(packageregistryv1.PackageArtifact{})},
	{FactKindPackageRegistryVulnerabilityHint, "package_registry.vulnerability_hint.v1.schema.json", reflect.TypeOf(packageregistryv1.VulnerabilityHint{})},
	{FactKindPackageRegistryRegistryEvent, "package_registry.registry_event.v1.schema.json", reflect.TypeOf(packageregistryv1.RegistryEvent{})},
	{FactKindPackageRegistryRepositoryHosting, "package_registry.repository_hosting.v1.schema.json", reflect.TypeOf(packageregistryv1.RepositoryHosting{})},
	{FactKindPackageRegistryWarning, "package_registry.warning.v1.schema.json", reflect.TypeOf(packageregistryv1.Warning{})},
	{FactKindSBOMDocument, "sbom.document.v1.schema.json", reflect.TypeOf(sbomv1.Document{})},
	{FactKindSBOMComponent, "sbom.component.v1.schema.json", reflect.TypeOf(sbomv1.Component{})},
	{FactKindSBOMDependencyRelationship, "sbom.dependency_relationship.v1.schema.json", reflect.TypeOf(sbomv1.DependencyRelationship{})},
	{FactKindSBOMExternalReference, "sbom.external_reference.v1.schema.json", reflect.TypeOf(sbomv1.ExternalReference{})},
	{FactKindSBOMWarning, "sbom.warning.v1.schema.json", reflect.TypeOf(sbomv1.Warning{})},
	{FactKindAttestationStatement, "attestation.statement.v1.schema.json", reflect.TypeOf(sbomv1.Statement{})},
	{FactKindAttestationSignatureVerification, "attestation.signature_verification.v1.schema.json", reflect.TypeOf(sbomv1.SignatureVerification{})},
	{FactKindAttestationSLSAProvenance, "attestation.slsa_provenance.v1.schema.json", reflect.TypeOf(sbomv1.SLSAProvenance{})},
	{FactKindScannerWorkerAnalysis, "scanner_worker.analysis.v1.schema.json", reflect.TypeOf(scannerworkerv1.Analysis{})},
	{FactKindScannerWorkerWarning, "scanner_worker.warning.v1.schema.json", reflect.TypeOf(scannerworkerv1.Warning{})},
	{FactKindVulnerabilityCVE, "vulnerability.cve.v1.schema.json", reflect.TypeOf(vulnerabilityv1.CVE{})},
	{FactKindVulnerabilityAffectedPackage, "vulnerability.affected_package.v1.schema.json", reflect.TypeOf(vulnerabilityv1.AffectedPackage{})},
	{FactKindVulnerabilityAffectedProduct, "vulnerability.affected_product.v1.schema.json", reflect.TypeOf(vulnerabilityv1.AffectedProduct{})},
	{FactKindVulnerabilityOSPackage, "vulnerability.os_package.v1.schema.json", reflect.TypeOf(vulnerabilityv1.OSPackage{})},
	{FactKindVulnerabilityEPSSScore, "vulnerability.epss_score.v1.schema.json", reflect.TypeOf(vulnerabilityv1.EPSSScore{})},
	{FactKindVulnerabilityKnownExploited, "vulnerability.known_exploited.v1.schema.json", reflect.TypeOf(vulnerabilityv1.KnownExploited{})},
	{FactKindVulnerabilityGoModuleEvidence, "vulnerability.go_module_evidence.v1.schema.json", reflect.TypeOf(vulnerabilityv1.GoModuleEvidence{})},
	{FactKindVulnerabilityGoCallReachability, "vulnerability.go_call_reachability.v1.schema.json", reflect.TypeOf(vulnerabilityv1.GoCallReachability{})},
	{FactKindVulnerabilityReference, "vulnerability.reference.v1.schema.json", reflect.TypeOf(vulnerabilityv1.Reference{})},
	{FactKindVulnerabilitySourceSnapshot, "vulnerability.source_snapshot.v1.schema.json", reflect.TypeOf(vulnerabilityv1.SourceSnapshot{})},
	{FactKindVulnerabilitySuppression, "vulnerability.suppression.v1.schema.json", reflect.TypeOf(vulnerabilitysuppressionv1.Suppression{})},
	{FactKindCodegraphFile, "file.v1.schema.json", reflect.TypeOf(codegraphv1.File{})},
	{FactKindCodegraphRepository, "repository.v1.schema.json", reflect.TypeOf(codegraphv1.Repository{})},
	{FactKindCodeDataflowScanned, "code_dataflow_scanned.v1.schema.json", reflect.TypeOf(codedataflowv1.DataflowScanned{})},
	{FactKindCodeDataflowFunction, "code_dataflow_function.v1.schema.json", reflect.TypeOf(codedataflowv1.DataflowFunction{})},
	{FactKindCodeFunctionSummary, "code_function_summary.v1.schema.json", reflect.TypeOf(codedataflowv1.FunctionSummary{})},
	{FactKindCodeFunctionSource, "code_function_source.v1.schema.json", reflect.TypeOf(codedataflowv1.FunctionSource{})},
	{FactKindCodeTaintEvidence, "code_taint_evidence.v1.schema.json", reflect.TypeOf(codedataflowv1.TaintEvidence{})},
	{FactKindCodeInterprocEvidence, "code_interproc_evidence.v1.schema.json", reflect.TypeOf(codedataflowv1.InterprocEvidence{})},
	{FactKindCICDRun, "ci.run.v1.schema.json", reflect.TypeOf(cicdrunv1.Run{})},
	{FactKindCICDArtifact, "ci.artifact.v1.schema.json", reflect.TypeOf(cicdrunv1.Artifact{})},
	{FactKindCICDEnvironmentObservation, "ci.environment_observation.v1.schema.json", reflect.TypeOf(cicdrunv1.EnvironmentObservation{})},
	{FactKindCICDDeploymentEvent, "ci.deployment_event.v1.schema.json", reflect.TypeOf(cicdrunv1.DeploymentEvent{})},
	{FactKindCICDTriggerEdge, "ci.trigger_edge.v1.schema.json", reflect.TypeOf(cicdrunv1.TriggerEdge{})},
	{FactKindCICDStep, "ci.step.v1.schema.json", reflect.TypeOf(cicdrunv1.Step{})},
	{FactKindCICDWorkflowImageEvidence, "ci.workflow_image_evidence.v1.schema.json", reflect.TypeOf(cicdrunv1.WorkflowImageEvidence{})},
	{FactKindVaultAuthRole, "vault_auth_role.v1.schema.json", reflect.TypeOf(secretsiamv1.VaultAuthRole{})},
	{FactKindVaultACLPolicy, "vault_acl_policy.v1.schema.json", reflect.TypeOf(secretsiamv1.VaultACLPolicy{})},
	{FactKindVaultKVMetadata, "vault_kv_metadata.v1.schema.json", reflect.TypeOf(secretsiamv1.VaultKVMetadata{})},
	{FactKindKubernetesServiceAccount, "k8s_service_account.v1.schema.json", reflect.TypeOf(secretsiamv1.KubernetesServiceAccount{})},
	{FactKindKubernetesWorkloadIdentityUse, "k8s_workload_identity_use.v1.schema.json", reflect.TypeOf(secretsiamv1.KubernetesWorkloadIdentityUse{})},
	{FactKindEKSIRSAAnnotation, "eks_irsa_annotation.v1.schema.json", reflect.TypeOf(secretsiamv1.EKSIRSAAnnotation{})},
	{FactKindEKSPodIdentityAssociation, "eks_pod_identity_association.v1.schema.json", reflect.TypeOf(secretsiamv1.EKSPodIdentityAssociation{})},
	{FactKindKubernetesGCPWorkloadIdentityBinding, "k8s_gcp_workload_identity_binding.v1.schema.json", reflect.TypeOf(secretsiamv1.KubernetesGCPWorkloadIdentityBinding{})},
	{FactKindKubernetesRBACRole, "k8s_rbac_role.v1.schema.json", reflect.TypeOf(secretsiamv1.KubernetesRBACRole{})},
	{FactKindKubernetesRBACBinding, "k8s_rbac_binding.v1.schema.json", reflect.TypeOf(secretsiamv1.KubernetesRBACBinding{})},
	{FactKindKubernetesServiceAccountTokenPosture, "k8s_service_account_token_posture.v1.schema.json", reflect.TypeOf(secretsiamv1.KubernetesServiceAccountTokenPosture{})},
	{FactKindVaultAuthMount, "vault_auth_mount.v1.schema.json", reflect.TypeOf(secretsiamv1.VaultAuthMount{})},
	{FactKindVaultIdentityEntity, "vault_identity_entity.v1.schema.json", reflect.TypeOf(secretsiamv1.VaultIdentityEntity{})},
	{FactKindVaultIdentityAlias, "vault_identity_alias.v1.schema.json", reflect.TypeOf(secretsiamv1.VaultIdentityAlias{})},
	{FactKindVaultSecretEngineMount, "vault_secret_engine_mount.v1.schema.json", reflect.TypeOf(secretsiamv1.VaultSecretEngineMount{})},
	{FactKindSecretsIAMCoverageWarning, "secrets_iam_coverage_warning.v1.schema.json", reflect.TypeOf(secretsiamv1.CoverageWarning{})},
	{FactKindWorkItemRecord, "work_item.record.v1.schema.json", reflect.TypeOf(workitemv1.WorkItemRecord{})},
	{FactKindWorkItemTransition, "work_item.transition.v1.schema.json", reflect.TypeOf(workitemv1.WorkItemTransition{})},
	{FactKindWorkItemExternalLink, "work_item.external_link.v1.schema.json", reflect.TypeOf(workitemv1.WorkItemExternalLink{})},
	{FactKindWorkItemProjectMetadata, "work_item.project_metadata.v1.schema.json", reflect.TypeOf(workitemv1.WorkItemProjectMetadata{})},
	{FactKindWorkItemIssueTypeMetadata, "work_item.issue_type_metadata.v1.schema.json", reflect.TypeOf(workitemv1.WorkItemIssueTypeMetadata{})},
	{FactKindWorkItemStatusMetadata, "work_item.status_metadata.v1.schema.json", reflect.TypeOf(workitemv1.WorkItemStatusMetadata{})},
	{FactKindWorkItemWorkflowMetadata, "work_item.workflow_metadata.v1.schema.json", reflect.TypeOf(workitemv1.WorkItemWorkflowMetadata{})},
	{FactKindWorkItemFieldMetadata, "work_item.field_metadata.v1.schema.json", reflect.TypeOf(workitemv1.WorkItemFieldMetadata{})},
	{FactKindWorkItemMetadataWarning, "work_item.metadata_warning.v1.schema.json", reflect.TypeOf(workitemv1.WorkItemMetadataWarning{})},
	{FactKindSecurityAlertRepositoryAlert, "security_alert.repository_alert.v1.schema.json", reflect.TypeOf(securityalertv1.RepositoryAlert{})},
	{FactKindReducerSupplyChainImpactFinding, "reducer_supply_chain_impact_finding.v1.schema.json", reflect.TypeOf(reducerderivedv1.SupplyChainImpactFinding{})},
	{FactKindReducerAWSCloudRuntimeDriftFinding, "reducer_aws_cloud_runtime_drift_finding.v1.schema.json", reflect.TypeOf(reducerderivedv1.AWSCloudRuntimeDriftFinding{})},
	{FactKindReducerMultiCloudRuntimeDriftFinding, "reducer_multi_cloud_runtime_drift_finding.v1.schema.json", reflect.TypeOf(reducerderivedv1.MultiCloudRuntimeDriftFinding{})},
	{FactKindReducerTerraformConfigStateDriftFinding, "reducer_terraform_config_state_drift_finding.v1.schema.json", reflect.TypeOf(reducerderivedv1.TerraformConfigStateDriftFinding{})},
	{FactKindReducerCodeDriftedFinding, "reducer_code_drifted_finding.v1.schema.json", reflect.TypeOf(reducerderivedv1.CodeDriftedFinding{})},
	{FactKindReducerPackageOwnershipCorrelation, "reducer_package_ownership_correlation.v1.schema.json", reflect.TypeOf(reducerderivedv1.PackageOwnershipCorrelation{})},
	{FactKindReducerPackageConsumptionCorrelation, "reducer_package_consumption_correlation.v1.schema.json", reflect.TypeOf(reducerderivedv1.PackageConsumptionCorrelation{})},
	{FactKindReducerPackagePublicationCorrelation, "reducer_package_publication_correlation.v1.schema.json", reflect.TypeOf(reducerderivedv1.PackagePublicationCorrelation{})},
	{FactKindObservabilityDeclaredFolder, "observability.declared_folder.v1.schema.json", reflect.TypeOf(observabilityv1.DeclaredFolder{})},
	{FactKindObservabilityDeclaredDashboard, "observability.declared_dashboard.v1.schema.json", reflect.TypeOf(observabilityv1.DeclaredDashboard{})},
	{FactKindObservabilityDeclaredDatasource, "observability.declared_datasource.v1.schema.json", reflect.TypeOf(observabilityv1.DeclaredDatasource{})},
	{FactKindObservabilityDeclaredAlertRule, "observability.declared_alert_rule.v1.schema.json", reflect.TypeOf(observabilityv1.DeclaredAlertRule{})},
	{FactKindObservabilityDeclaredScrapeConfig, "observability.declared_scrape_config.v1.schema.json", reflect.TypeOf(observabilityv1.DeclaredScrapeConfig{})},
	{FactKindObservabilityDeclaredMetricRule, "observability.declared_metric_rule.v1.schema.json", reflect.TypeOf(observabilityv1.DeclaredMetricRule{})},
	{FactKindObservabilityDeclaredMetricRoute, "observability.declared_metric_route.v1.schema.json", reflect.TypeOf(observabilityv1.DeclaredMetricRoute{})},
	{FactKindObservabilityDeclaredLogRoute, "observability.declared_log_route.v1.schema.json", reflect.TypeOf(observabilityv1.DeclaredLogRoute{})},
	{FactKindObservabilityDeclaredTraceRoute, "observability.declared_trace_route.v1.schema.json", reflect.TypeOf(observabilityv1.DeclaredTraceRoute{})},
	{FactKindObservabilityAppliedResource, "observability.applied_resource.v1.schema.json", reflect.TypeOf(observabilityv1.AppliedResource{})},
	{FactKindObservabilityAppliedSyncState, "observability.applied_sync_state.v1.schema.json", reflect.TypeOf(observabilityv1.AppliedSyncState{})},
	{FactKindObservabilityObservedDashboard, "observability.observed_dashboard.v1.schema.json", reflect.TypeOf(observabilityv1.ObservedDashboard{})},
	{FactKindObservabilityObservedTarget, "observability.observed_target.v1.schema.json", reflect.TypeOf(observabilityv1.ObservedTarget{})},
	{FactKindObservabilityObservedRule, "observability.observed_rule.v1.schema.json", reflect.TypeOf(observabilityv1.ObservedRule{})},
	{FactKindObservabilityObservedLogSignal, "observability.observed_log_signal.v1.schema.json", reflect.TypeOf(observabilityv1.ObservedLogSignal{})},
	{FactKindObservabilityObservedTraceSignal, "observability.observed_trace_signal.v1.schema.json", reflect.TypeOf(observabilityv1.ObservedTraceSignal{})},
	{FactKindObservabilityCoverageWarning, "observability.coverage_warning.v1.schema.json", reflect.TypeOf(observabilityv1.CoverageWarning{})},
	{FactKindObservabilitySourceInstance, "observability.source_instance.v1.schema.json", reflect.TypeOf(observabilityv1.SourceInstance{})},
	{FactKindDocumentationSource, "documentation_source.v1.schema.json", reflect.TypeOf(documentationv1.Source{})},
	{FactKindDocumentationDocument, "documentation_document.v1.schema.json", reflect.TypeOf(documentationv1.Document{})},
	{FactKindDocumentationSection, "documentation_section.v1.schema.json", reflect.TypeOf(documentationv1.Section{})},
	{FactKindDocumentationLink, "documentation_link.v1.schema.json", reflect.TypeOf(documentationv1.Link{})},
	{FactKindDocumentationEntityMention, "documentation_entity_mention.v1.schema.json", reflect.TypeOf(documentationv1.EntityMention{})},
	{FactKindDocumentationClaimCandidate, "documentation_claim_candidate.v1.schema.json", reflect.TypeOf(documentationv1.ClaimCandidate{})},
	{FactKindDocumentationFinding, "documentation_finding.v1.schema.json", reflect.TypeOf(documentationv1.Finding{})},
	{FactKindDocumentationEvidencePacket, "documentation_evidence_packet.v1.schema.json", reflect.TypeOf(documentationv1.EvidencePacket{})},
	{FactKindSemanticDocumentationObservation, "semantic.documentation_observation.v1.schema.json", reflect.TypeOf(semanticv1.DocumentationObservation{})},
	{FactKindSemanticCodeHint, "semantic.code_hint.v1.schema.json", reflect.TypeOf(semanticv1.CodeHint{})},
	{FactKindServiceCatalogEntity, "service_catalog.entity.v1.schema.json", reflect.TypeOf(servicecatalogv1.Entity{})},
	{FactKindServiceCatalogOwnership, "service_catalog.ownership.v1.schema.json", reflect.TypeOf(servicecatalogv1.Ownership{})},
	{FactKindServiceCatalogRepositoryLink, "service_catalog.repository_link.v1.schema.json", reflect.TypeOf(servicecatalogv1.RepositoryLink{})},
	{FactKindServiceCatalogOperationalLink, "service_catalog.operational_link.v1.schema.json", reflect.TypeOf(servicecatalogv1.OperationalLink{})},
	{FactKindSubmodulePin, "submodule.pin.v1.schema.json", reflect.TypeOf(submodulev1.Pin{})},
	{FactKindCodeownersOwnership, "codeowners.ownership.v1.schema.json", reflect.TypeOf(codeownersv1.Ownership{})},
	{FactKindRefOverlayBranchObservation, "ref_overlay.branch_observation.v1.schema.json", reflect.TypeOf(refOverlayV1.BranchObservation{})},
	{FactKindConfigKeyDivergence, "config_key_divergence.v1.schema.json", reflect.TypeOf(refOverlayV1.ConfigKeyDivergence{})},
}

// TestPayloadContractsCoverAllSchemas fails if the payloadContracts registry
// and the checked-in schema/ directory disagree about which fact kinds exist:
// a schema file with no registry row (a kind added without wiring its drift
// coverage) or a registry row naming a missing schema file. This is the guard
// that keeps "add a kind" from silently skipping the single-source-of-truth
// checks the other tests enforce.
