## operator-cli

### plan
Tha plan command works by getting the cluster CRs in the current cluster context, updating the kops state, and generating the terraforms files, similar with what the reconciliaton does. The context to use this is when the cluster CRs are already deployed in the cluster paused. The CLI needs two parameters: the cluster CR name and the namespace where it belongs.

Usage:
`go run cli/main.go plan --name infra-test.eu-central-1.k8s.tfgco.com --namespace k8s-infra-test-eu-central-1-general-test`

Next steps:
- Kops should have a way to calculate the diff without updating the kops state, we should start doing that
- This script requires that the cluster CRs are already deployed, we should think in a way to do this using files to integrate better with pipeline

