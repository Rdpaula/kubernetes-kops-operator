/*
Copyright 2021.

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

package controlplane

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	controlplanev1alpha1 "github.com/topfreegames/kubernetes-kops-operator/apis/controlplane/v1alpha1"
	"github.com/topfreegames/kubernetes-kops-operator/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/assets"

	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/pkg/commands"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/cluster-api/util/predicates"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// KopsControlPlaneReconciler reconciles a KopsControlPlane object
type KopsControlPlaneReconciler struct {
	client.Client
	Scheme               *runtime.Scheme
	kopsClientset        simple.Clientset
	log                  logr.Logger
	PopulateClusterSpec  func(cluster *kopsapi.Cluster, kopsClientset simple.Clientset) (*kopsapi.Cluster, error)
	CreateCloudResources func(kopsClientset simple.Clientset, ctx context.Context, kopsCluster *kopsapi.Cluster, configBase string) error
	GetClusterStatus     func(cluster *kopsapi.Cluster) (*kopsapi.ClusterStatus, error)
}

// PopulateClusterSpec populates the full cluster spec with some values it fetchs from provider
func PopulateClusterSpec(cluster *kopsapi.Cluster, kopsClientset simple.Clientset) (*kopsapi.Cluster, error) {
	cloud, err := cloudup.BuildCloud(cluster)
	if err != nil {
		return nil, err
	}

	err = cloudup.PerformAssignments(cluster, cloud)
	if err != nil {
		return nil, err
	}

	assetBuilder := assets.NewAssetBuilder(cluster, "")
	fullCluster, err := cloudup.PopulateClusterSpec(kopsClientset, cluster, cloud, assetBuilder)
	if err != nil {
		return nil, err
	}

	return fullCluster, nil
}

// CreateCloudResources renders the terraform files and effectively apply them in the cloud provider
func CreateCloudResources(kopsClientset simple.Clientset, ctx context.Context, kopsCluster *kopsapi.Cluster, configBase string) error {
	s3Bucket, err := utils.GetBucketName(configBase)
	if err != nil {
		return err
	}

	terraformOutputDir := fmt.Sprintf("/tmp/%s", kopsCluster.Name)

	cloud, err := cloudup.BuildCloud(kopsCluster)
	if err != nil {
		return err
	}

	applyCmd := &cloudup.ApplyClusterCmd{
		Cloud:              cloud,
		Clientset:          kopsClientset,
		Cluster:            kopsCluster,
		DryRun:             true,
		AllowKopsDowngrade: false,
		OutDir:             terraformOutputDir,
		TargetName:         "terraform",
	}

	if err := applyCmd.Run(ctx); err != nil {
		return err
	}

	err = utils.CreateTerraformBackendFile(s3Bucket, kopsCluster.Name, terraformOutputDir)
	if err != nil {
		return err
	}

	err = utils.ApplyTerraform(ctx, terraformOutputDir)
	if err != nil {
		return err
	}
	return nil

}

// GetClusterStatus retrieve the kops cluster status from the cloud provider
func GetClusterStatus(kopsCluster *kopsapi.Cluster) (*kopsapi.ClusterStatus, error) {
	statusDiscovery := &commands.CloudDiscoveryStatusStore{}
	status, err := statusDiscovery.FindClusterStatus(kopsCluster)
	if err != nil {
		return nil, err
	}
	return status, nil
}

// addSSHCredential creates a SSHCredential using the PublicKey retrieved from the KopsControlPlane
func (r *KopsControlPlaneReconciler) addSSHCredential(cluster *kopsapi.Cluster, SSHPublicKey string) error {
	sshCredential := kopsapi.SSHCredential{
		Spec: kopsapi.SSHCredentialSpec{
			PublicKey: SSHPublicKey,
		},
	}

	sshCredentialStore, err := r.kopsClientset.SSHCredentialStore(cluster)
	if err != nil {
		return err
	}
	sshKeyArr := []byte(sshCredential.Spec.PublicKey)
	err = sshCredentialStore.AddSSHPublicKey("admin", sshKeyArr)
	if err != nil {
		return err
	}

	r.log.Info("Added ssh credential")

	return nil
}

// updateKopsState creates or updates the kops state in the remote storage
func (r *KopsControlPlaneReconciler) updateKopsState(ctx context.Context, kopsCluster *kopsapi.Cluster, SSHPublicKey string) error {
	oldCluster, _ := r.kopsClientset.GetCluster(ctx, kopsCluster.Name)
	if oldCluster != nil {
		status, err := r.GetClusterStatus(oldCluster)
		if err != nil {
			return err
		}
		r.kopsClientset.UpdateCluster(ctx, kopsCluster, status)
		r.log.Info(fmt.Sprintf("updated kops state for cluster %s", kopsCluster.ObjectMeta.Name))
		return nil
	}

	_, err := r.kopsClientset.CreateCluster(ctx, kopsCluster)
	if err != nil {
		return err
	}

	err = r.addSSHCredential(kopsCluster, SSHPublicKey)
	if err != nil {
		return err
	}

	r.log.Info(fmt.Sprintf("created kops state for cluster %s", kopsCluster.ObjectMeta.Name))

	return nil
}

//+kubebuilder:rbac:groups=controlplane.cluster.x-k8s.io,resources=kopscontrolplanes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=controlplane.cluster.x-k8s.io,resources=kopscontrolplanes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=controlplane.cluster.x-k8s.io,resources=kopscontrolplanes/finalizers,verbs=update
func (r *KopsControlPlaneReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, rerr error) {
	r.log = ctrl.LoggerFrom(ctx)

	applicableConditions := []clusterv1.ConditionType{
		controlplanev1alpha1.KopsControlPlaneReadyCondition,
	}

	var kopsControlPlane controlplanev1alpha1.KopsControlPlane
	if err := r.Get(ctx, req.NamespacedName, &kopsControlPlane); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Initialize the patch helper.
	patchHelper, err := patch.NewHelper(&kopsControlPlane, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Attempt to Patch the KopsControlPlane object and status after each reconciliation if no error occurs.
	defer func() {
		conditions.SetSummary(&kopsControlPlane,
			conditions.WithConditions(
				applicableConditions...,
			),
			conditions.WithStepCounter(),
		)

		err = patchHelper.Patch(ctx, &kopsControlPlane,
			patch.WithOwnedConditions{
				Conditions: applicableConditions,
			},
		)
		if err != nil {
			r.log.Error(rerr, "Failed to patch kopsControlPlane")
			if rerr == nil {
				rerr = err
			}
		}
	}()

	controllerutil.AddFinalizer(&kopsControlPlane, controlplanev1alpha1.KopsControlPlaneFinalizer)
	err = patchHelper.Patch(ctx, &kopsControlPlane,
		patch.WithOwnedConditions{
			Conditions: applicableConditions,
		},
	)
	if err != nil {
		return ctrl.Result{}, err
	}

	kopsClientset, err := utils.GetKopsClientset(kopsControlPlane.Spec.KopsClusterSpec.ConfigBase)
	if err != nil {
		return ctrl.Result{}, err
	}

	r.kopsClientset = kopsClientset

	kopsCluster := &kopsapi.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: kopsControlPlane.ObjectMeta.Labels[kopsapi.LabelClusterName],
		},
		Spec: kopsControlPlane.Spec.KopsClusterSpec,
	}

	fullCluster, err := r.PopulateClusterSpec(kopsCluster, r.kopsClientset)
	if err != nil {
		return ctrl.Result{}, err
	}

	err = r.updateKopsState(ctx, fullCluster, kopsControlPlane.Spec.SSHPublicKey)
	if err != nil {
		r.log.Error(err, fmt.Sprintf("failed to create cluster: %v", err))
		return ctrl.Result{}, err
	}

	err = r.CreateCloudResources(r.kopsClientset, ctx, kopsCluster, fullCluster.Spec.ConfigBase)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KopsControlPlaneReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&controlplanev1alpha1.KopsControlPlane{}).
		WithEventFilter(predicates.ResourceNotPaused(ctrl.LoggerFrom(ctx))).
		Watches(
			&source.Kind{Type: &clusterv1.Cluster{}},
			handler.EnqueueRequestsFromMapFunc(clusterToInfrastructureMapFunc),
		).
		Complete(r)
}

func clusterToInfrastructureMapFunc(o client.Object) []ctrl.Request {
	c, ok := o.(*clusterv1.Cluster)
	if !ok {
		panic(fmt.Sprintf("Expected a Cluster but got a %T", o))
	}

	result := []ctrl.Request{}
	if c.Spec.InfrastructureRef != nil && c.Spec.InfrastructureRef.GroupVersionKind() == controlplanev1alpha1.GroupVersion.WithKind("KopsControlPlane") {
		name := client.ObjectKey{Namespace: c.Spec.InfrastructureRef.Namespace, Name: c.Spec.InfrastructureRef.Name}
		result = append(result, ctrl.Request{NamespacedName: name})
	}

	return result
}
