package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"runtime"

	"github.com/jenkins-x/sso-operator/pkg/dex"
	"github.com/jenkins-x/sso-operator/pkg/kubernetes"
	"github.com/jenkins-x/sso-operator/pkg/operator"
	sdk "github.com/operator-framework/operator-sdk/pkg/sdk"
	sdkVersion "github.com/operator-framework/operator-sdk/version"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	watchNamespaceEnv    = "WATCH_NAMESPACE"
	operatorNamespaceEnv = "OPERATOR_NAMESPACE"
	port                 = "8080"
)

// OperatorOptions holds the command options for SSO operator
type OperatorOptions struct {
	Namespace          string
	WatchNamespace     string
	DexGrpcHostAndPort string
	DexGrpcClientCrt   string
	DexGrpcClientKey   string
	DexGrpcClientCA    string
	ClusterRoleName    string
}

func printVersion(namespace string, watchNamespace string) {
	logrus.Infof("Go Version: %s", runtime.Version())
	logrus.Infof("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)
	logrus.Infof("operator-sdk Version: %v", sdkVersion.Version)
	logrus.Infof("operator namespace: %s", namespace)
	if watchNamespace == "" {
		logrus.Info("operator watching entire cluster")
	} else {
		logrus.Infof("operator watching namespace: %s", watchNamespace)
	}
}

func handleLiveness() {
	logrus.Infof("Liveness probe listening on: %s", port)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logrus.Debug("ping")
	})
	err := http.ListenAndServe(":"+port, nil) // #nosec
	if err != nil {
		logrus.Errorf("failed to start health probe: %v\n", err)
	}
}

// Run starts the SSO operator
func (o *OperatorOptions) Run() {
	namespace := o.Namespace
	if namespace == "" {
		namespace = os.Getenv(operatorNamespaceEnv)
	}
	watchNamespace := o.WatchNamespace
	if watchNamespace == "" {
		watchNamespace = os.Getenv(watchNamespaceEnv)
	}
	printVersion(namespace, watchNamespace)

	// validate the command line options
	err := o.Validate()
	if err != nil {
		logrus.Errorf("invalid options: %v", err)
		os.Exit(2)
	}

	opts := &dex.Options{
		HostAndPort: o.DexGrpcHostAndPort,
		ClientCrt:   o.DexGrpcClientCrt,
		ClientKey:   o.DexGrpcClientKey,
		ClientCA:    o.DexGrpcClientCA,
	}
	dexClient, err := dex.NewClient(opts)
	if err != nil {
		logrus.Errorf("failed to crate dex client: %v", err)
		os.Exit(2)
	}

	logrus.Infof("Connected to Dex gRPC server: %s", o.DexGrpcHostAndPort)

	// Register the CRDs
	apiclient, err := kubernetes.GetAPIExtensionsClient()
	if err != nil {
		logrus.Errorf("failed to register the k8s API extensions client: %v", err)
		os.Exit(2)
	}
	err = kubernetes.RegisterSSOCRD(apiclient)
	if err != nil {
		logrus.Errorf("failed to register the SSO CRD: %v", err)
		os.Exit(2)
	}

	// configure the operator
	sdk.Watch("jenkins.io/v1", "SSO", watchNamespace, 5)
	handler, err := operator.NewHandler(dexClient, namespace, o.ClusterRoleName)
	if err != nil {
		logrus.Errorf("failed to create the operator handler: %v", err)
		os.Exit(2)
	}
	sdk.Handle(handler)

	// start the health probe
	go handleLiveness()

	// start the operator
	sdk.Run(context.TODO())
}

// Validate validates the provided command options
func (o *OperatorOptions) Validate() error {
	if o.DexGrpcHostAndPort == "" {
		return errors.New("dex gRPC server host and port is empty")
	}
	if _, err := os.Stat(o.DexGrpcClientCrt); os.IsNotExist(err) {
		return fmt.Errorf("provided dex gRPC client cert file '%s' does not exist", o.DexGrpcClientCrt)
	}

	if _, err := os.Stat(o.DexGrpcClientKey); os.IsNotExist(err) {
		return fmt.Errorf("provided dex gRPC client key file '%s' does not exists", o.DexGrpcClientKey)
	}

	if _, err := os.Stat(o.DexGrpcClientCA); os.IsNotExist(err) {
		return fmt.Errorf("provided dex gRPC CA cert file '%s' does not exists", o.DexGrpcClientCA)
	}

	return nil
}

func commandRoot() *cobra.Command {
	options := &OperatorOptions{}

	rootCmd := &cobra.Command{
		Use: "sso-operator",
		Run: func(cmd *cobra.Command, args []string) {
			options.Run()
		},
	}

	rootCmd.Flags().StringVarP(&options.Namespace, "namespace", "n", "", "Namespace where the operator where the operator is deployed")
	rootCmd.Flags().StringVarP(&options.WatchNamespace, "watch-namespace", "", "", "Namespace where the operator will watch for resources (leave empty to watch the entire cluster)")
	rootCmd.Flags().StringVarP(&options.DexGrpcHostAndPort, "dex-grpc-host-port", "", "", "Host and port of Dex gRPC server")
	rootCmd.Flags().StringVarP(&options.DexGrpcClientCrt, "dex-grpc-client-crt", "", "", "Certificate for Dex gRPC client")
	rootCmd.Flags().StringVarP(&options.DexGrpcClientKey, "dex-grpc-client-key", "", "", "Key for Dex gRPC client")
	rootCmd.Flags().StringVarP(&options.DexGrpcClientCA, "dex-grpc-client-ca", "", "", "CA certificate for Dex gRPC client")
	rootCmd.Flags().StringVarP(&options.ClusterRoleName, "cluster-role-name", "", "", "Cluster role name which has the required permissions for operator")

	return rootCmd
}

func main() {
	if err := commandRoot().Execute(); err != nil {
		logrus.Error(err)
		os.Exit(2)
	}
}
