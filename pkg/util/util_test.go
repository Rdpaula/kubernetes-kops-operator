package util

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubectl/pkg/scheme"
	clusterv1betav1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestGetClusterByName(t *testing.T) {
	testCases := []struct {
		description   string
		clusters      []client.Object
		expectedError bool
	}{
		{
			description:   "Should successfully return cluster",
			expectedError: false,
			clusters: []client.Object{
				newCluster("test-cluster", "", metav1.NamespaceDefault),
			},
		},
		{
			description:   "Cluster don't exist, should return error",
			expectedError: true,
			clusters:      nil,
		},
	}

	RegisterFailHandler(Fail)
	g := NewWithT(t)
	ctx := context.TODO()

	err := clusterv1betav1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			fakeClient := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(tc.clusters...).Build()
			cluster, err := GetClusterByName(ctx, fakeClient, metav1.NamespaceDefault, "test-cluster.k8s.cluster")
			if !tc.expectedError {
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(cluster).NotTo(BeNil())
			} else {
				g.Expect(err).To(HaveOccurred())
			}

		})
	}
}

func TestSetAWSEnvFromKopsControlPlaneSecret(t *testing.T) {
	testCases := []struct {
		description   string
		k8sObjects    []client.Object
		expectedError bool
	}{
		{
			description:   "Should successfully set AWS envs",
			expectedError: false,
			k8sObjects: []client.Object{
				newAWSCredentialSecret("11111111-credential", "kubernetes-kops-operator-system"),
			},
		},
	}

	RegisterFailHandler(Fail)
	g := NewWithT(t)

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			awsCredentials := aws.Credentials{
				AccessKeyID:     "11111111-credential",
				SecretAccessKey: "kubernetes-kops-operator-system",
			}
			err := SetEnvVarsFromAWSCredentials(awsCredentials)
			if !tc.expectedError {
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(os.Getenv("AWS_ACCESS_KEY_ID")).To(Equal("11111111-credential"))
				g.Expect(os.Getenv("AWS_SECRET_ACCESS_KEY")).To(Equal("kubernetes-kops-operator-system"))
			} else {
				g.Expect(err).To(HaveOccurred())
			}

		})
	}
}

func TestGetAwsCredentialsFromKopsControlPlaneSecret(t *testing.T) {
	testCases := []struct {
		description           string
		k8sObjects            []client.Object
		expectedAwsCredential *aws.Credentials
		expectedError         bool
	}{
		{
			description:   "Should successfully set AWS envs",
			expectedError: false,
			k8sObjects: []client.Object{
				newAWSCredentialSecret("accessTest", "secretTest"),
			},
			expectedAwsCredential: &aws.Credentials{AccessKeyID: "accessTest", SecretAccessKey: "secretTest"},
		},
		{
			description:   "Should fail if can't get secret",
			expectedError: true,
			k8sObjects:    []client.Object{},
		},
	}

	RegisterFailHandler(Fail)
	g := NewWithT(t)
	ctx := context.TODO()

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			fakeClient := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(tc.k8sObjects...).Build()
			credential, err := GetAWSCredentialsFromKopsControlPlaneSecret(ctx, fakeClient, "11111111-credential", "kubernetes-kops-operator-system")
			if !tc.expectedError {
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(credential).To(Equal(tc.expectedAwsCredential))
			} else {
				g.Expect(err).To(HaveOccurred())
			}

		})
	}
}

func newAWSCredentialSecret(accessKey, secret string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "11111111-credential",
			Namespace: "kubernetes-kops-operator-system",
		},
		Data: map[string][]byte{
			"AccessKeyID":     []byte(accessKey),
			"SecretAccessKey": []byte(secret),
		},
	}
}

func newCluster(name, controlplane, namespace string) *clusterv1betav1.Cluster {
	return &clusterv1betav1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      fmt.Sprintf("%s.k8s.cluster", name),
		},
		Spec: clusterv1betav1.ClusterSpec{
			ControlPlaneRef: &corev1.ObjectReference{
				Name:      controlplane,
				Namespace: namespace,
				Kind:      "KopsControlPlane",
			},
		},
	}
}
