package main

import (
	"github.com/dbschenker/thundering-herd-scheduler/pkg/thunderingherdscheduling"
	"k8s.io/component-base/logs"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/cmd/kube-scheduler/app"
	"os"
)

const VERSION = "development"

func main() {
	klog.Info("Start custom scheduler implementing Thundering Herd Scheduling")
	klog.Infof("Version: %s", VERSION)
	klog.Info("Author: Kamil Krzywicki (kamil.krzywicki@dbschenker.com)")
	klog.Info("Author: Björn Wenzel (bjoern.wenzel@dbschenker.com)")

	command := app.NewSchedulerCommand(
		app.WithPlugin(thunderingherdscheduling.Name, thunderingherdscheduling.New),
	)

	logs.InitLogs()
	defer logs.FlushLogs()

	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
