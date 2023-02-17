package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/spf13/cobra"
	controlplanev1alpha1 "github.com/topfreegames/kubernetes-kops-operator/apis/controlplane/v1alpha1"
	infrastructurev1alpha1 "github.com/topfreegames/kubernetes-kops-operator/apis/infrastructure/v1alpha1"
	"github.com/topfreegames/kubernetes-kops-operator/controllers/controlplane"
	"github.com/topfreegames/kubernetes-kops-operator/utils"

	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/upup/pkg/fi"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "This command will show the Terraform plan",
	Long:  `This plan command will generate the terraform files in the same way as the controller and show the plan.`,
	Run: func(cmd *cobra.Command, args []string) {

		if len(clusterName) == 0 {
			panic("cluster not defined")
		}

		if len(namespace) == 0 {
			panic("namespace not defined")
		}

		scheme := runtime.NewScheme()

		utilruntime.Must(infrastructurev1alpha1.AddToScheme(scheme))

		utilruntime.Must(controlplanev1alpha1.AddToScheme(scheme))

		configPath := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		config, err := clientcmd.BuildConfigFromFlags("", configPath)
		if err != nil {
			panic(err)
		}

		k8sClient, err := client.New(config, client.Options{
			Scheme: scheme,
		})
		if err != nil {
			panic(err)
		}

		kopsControlPlane := &controlplanev1alpha1.KopsControlPlane{}
		key := client.ObjectKey{
			Namespace: namespace,
			Name:      clusterName,
		}

		if err := k8sClient.Get(context.TODO(), key, kopsControlPlane); err != nil {
			panic(err)
		}

		kopsClientset, err := utils.GetKopsClientset(kopsControlPlane.Spec.KopsClusterSpec.ConfigBase)
		if err != nil {
			panic(err)
		}

		kopsCluster := &kopsapi.Cluster{
			ObjectMeta: metav1.ObjectMeta{
				Name: kopsControlPlane.GetName(),
			},
			Spec: kopsControlPlane.Spec.KopsClusterSpec,
		}

		err = utils.ParseSpotinstFeatureflags(kopsControlPlane)
		if err != nil {
			panic(err)
		}
		cloud, err := utils.BuildCloud(kopsCluster)
		if err != nil {
			panic(err)
		}

		fullCluster, err := controlplane.PopulateClusterSpec(kopsCluster, kopsClientset, cloud)
		if err != nil {
			panic(err)
		}

		err = updateKopsCluster(context.TODO(), kopsClientset, fullCluster, kopsControlPlane.Spec.SSHPublicKey, cloud)
		if err != nil {
			panic(err)
		}

		if kopsControlPlane.Spec.KopsSecret != nil {
			secretStore, err := kopsClientset.SecretStore(kopsCluster)
			if err != nil {
				panic(err)
			}

			_ = utils.ReconcileKopsSecrets(context.TODO(), k8sClient, secretStore, kopsControlPlane, client.ObjectKey{
				Name:      kopsControlPlane.Spec.KopsSecret.Name,
				Namespace: kopsControlPlane.Spec.KopsSecret.Namespace,
			})
		}

		terraformOutputDir := fmt.Sprintf("/tmp/%s", kopsCluster.Name)

		// We are ignoring this part for now
		// var shouldIgnoreSG bool
		// if _, ok := owner.GetAnnotations()["kopscontrolplane.controlplane.wildlife.io/external-security-groups"]; ok {
		// 	shouldIgnoreSG = true
		// }
		shouldIgnoreSG := true

		err = controlplane.PrepareCloudResources(kopsClientset, k8sClient, context.TODO(), kopsCluster, kopsControlPlane, fullCluster.Spec.ConfigBase, terraformOutputDir, cloud, shouldIgnoreSG)
		if err != nil {
			panic(err)
		}

		const tfVersion = "1.3.4"

		tfPath := fmt.Sprintf("/tmp/%s_%s", product.Terraform.Name, tfVersion)

		_, err = os.Stat(tfPath)
		if os.IsNotExist(err) {
			err = os.Mkdir(tfPath, os.ModePerm)
			if err != nil {
				panic(err)
			}
		}

		if err != nil {
			panic(err)
		}

		installer := &releases.ExactVersion{
			Product:    product.Terraform,
			Version:    version.Must(version.NewVersion(tfVersion)),
			InstallDir: tfPath,
		}

		tfExecPath, err := installer.Install(context.TODO())
		if err != nil {
			panic(err)
		}

		err = PlanTerraform(context.TODO(), terraformOutputDir, tfExecPath)
		if err != nil {
			panic(err)
		}

	},
}

func updateKopsCluster(ctx context.Context, kopsClientset simple.Clientset, kopsCluster *kopsapi.Cluster, SSHPublicKey string, cloud fi.Cloud) error {
	kopsCluster, err := kopsClientset.GetCluster(ctx, kopsCluster.Name)
	if err != nil {
		return err
	}

	status, err := cloud.FindClusterStatus(kopsCluster)
	if err != nil {
		return err
	}
	_, err = kopsClientset.UpdateCluster(ctx, kopsCluster, status)
	if err != nil {
		return err
	}
	return nil
}

func PlanTerraform(ctx context.Context, workingDir, terraformExecPath string) error {

	tf, err := tfexec.NewTerraform(workingDir, terraformExecPath)
	if err != nil {
		return err
	}

	err = tf.Init(ctx, tfexec.Upgrade(true))
	if err != nil {
		return err
	}

	tf.SetStdout(os.Stdout)
	_, err = tf.Plan(ctx)
	if err != nil {
		return err
	}

	return nil
}

func init() {
	rootCmd.AddCommand(planCmd)
}
