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

func printVersion(namespace string) {
	logrus.Infof("Go Version: %s", runtime.Version())
	logrus.Infof("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)
	logrus.Infof("operator-sdk Version: %v", sdkVersion.Version)
	logrus.Infof("operator namespace: %s", namespace)
}

func handleHealth() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logrus.Debug("ping")
	})
	http.ListenAndServe(":8080", nil)
}

func main() {
	ns := os.Getenv(operatorNamespace)
	printVersion(ns)

	// configure the operator
	sdk.Watch("sso.jenkins.io/v1", "SSO", ns, 5)
	sdk.Handle(operator.NewHandler())

	// start the health probe
	go handleHealth()

	// start the operator
	sdk.Run(context.TODO())
}
