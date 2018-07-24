package main

import (
	"context"
	"net/http"
	"os"
	"runtime"

	"github.com/jenkins-x/sso-operator/pkg/operator"
	sdk "github.com/operator-framework/operator-sdk/pkg/sdk"
	sdkVersion "github.com/operator-framework/operator-sdk/version"

	"github.com/sirupsen/logrus"
)

const operatorNamespace = "OPERATOR_NAMESPACE"
const port = "8080"

func printVersion(namespace string) {
	logrus.Infof("Go Version: %s", runtime.Version())
	logrus.Infof("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)
	logrus.Infof("operator-sdk Version: %v", sdkVersion.Version)
	logrus.Infof("operator namespace: %s", namespace)
}

func handleLiveness() {
	logrus.Infof("Liveness probe listening on: %s", port)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logrus.Debug("ping")
	})
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		logrus.Errorf("failed to start health probe: %v\n", err)
	}
}

func main() {
	ns := os.Getenv(operatorNamespace)
	printVersion(ns)

	// configure the operator
	sdk.Watch("sso.jenkins.io/v1", "SSO", ns, 5)
	sdk.Handle(operator.NewHandler())

	// start the health probe
	go handleLiveness()

	// start the operator
	sdk.Run(context.TODO())
}
