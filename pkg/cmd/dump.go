/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/citrusframework/yaks/pkg/apis/yaks/v1alpha1"
	"github.com/citrusframework/yaks/pkg/client"
	"github.com/citrusframework/yaks/pkg/client/yaks/clientset/versioned"
	"github.com/citrusframework/yaks/pkg/cmd/config"
	cfg "github.com/citrusframework/yaks/pkg/config"
	"github.com/citrusframework/yaks/pkg/util"
	"github.com/citrusframework/yaks/pkg/util/kubernetes"

	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func newCmdDump(rootCmdOptions *RootCmdOptions) (*cobra.Command, *dumpCmdOptions) {
	options := dumpCmdOptions{
		RootCmdOptions: rootCmdOptions,
	}
	cmd := cobra.Command{
		Use:     "dump [filename]",
		Short:   "Dump the state of YAKS resources for a test",
		Long:    `Dump the state of currently used test resources in a namespace. If filename argument is missing, the output will be on Stdout`,
		PreRunE: decode(&options),
		RunE:    options.dump,
	}

	cmd.Flags().StringP("test", "t", "", "Name of test to dump")
	cmd.Flags().StringArrayP("includes", "i", nil, "Label selector to include pods when scraping pod logs.")
	cmd.Flags().IntP("lines", "l", 0, "Number of pod log lines to print")
	return &cmd, &options
}

type dumpCmdOptions struct {
	*RootCmdOptions
	Test     string   `mapstructure:"test"`
	Includes []string `mapstructure:"includes"`
	Lines    int      `mapstructure:"lines"`
}

func (o *dumpCmdOptions) dump(cmd *cobra.Command, args []string) error {
	c, err := o.GetCmdClient()
	if err != nil {
		return err
	}

	createDump := func(out io.Writer) error {
		if o.Test != "" {
			return dumpTest(o.Context, c, o.Test, o.Namespace, out, o.Lines, o.Includes)
		}

		return dumpAll(o.Context, c, o.Namespace, out, o.Lines, o.Includes)
	}

	if len(args) == 1 {
		return util.WithFile(args[0], os.O_RDWR|os.O_CREATE, 0o644, createDump)
	}

	return createDump(cmd.OutOrStdout())
}

func dumpTest(ctx context.Context, c client.Client, name string, namespace string, out io.Writer, logLines int, includes []string) error {
	yaksClient, err := versioned.NewForConfig(c.GetConfig())
	if err != nil {
		return err
	}

	test, err := yaksClient.YaksV1alpha1().Tests(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "Found test %s:\n", test.Name)
	err = printObject(test, out)
	if err != nil {
		return err
	}

	operatorNamespace := namespace
	if test.Labels[cfg.OperatorLabel] != "" {
		operatorNamespace = test.Labels[cfg.OperatorLabel]
	}

	err = dumpOperator(ctx, c, yaksClient, operatorNamespace, out, logLines)
	if err != nil {
		return err
	}

	testLabelSelector := metav1.ListOptions{}
	if len(test.Status.TestID) > 0 {
		testLabelSelector.LabelSelector = fmt.Sprintf("%s=%s", v1alpha1.TestIDLabel, test.Status.TestID)
	} else {
		testLabelSelector.LabelSelector = fmt.Sprintf("%s=%s", v1alpha1.TestLabel, test.Name)
	}

	cms, err := c.CoreV1().ConfigMaps(namespace).List(ctx, testLabelSelector)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "Found %d configmap(s):\n", len(cms.Items))
	for i := range cms.Items {
		err = printObject(&cms.Items[i], out)
		if err != nil {
			return err
		}
	}

	jobs, err := c.BatchV1().Jobs(namespace).List(ctx, testLabelSelector)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "Found %d job(s):\n", len(jobs.Items))
	for i := range jobs.Items {
		err = printObject(&jobs.Items[i], out)
		if err != nil {
			return err
		}
	}

	err = dumpPods(ctx, c, namespace, out, testLabelSelector, logLines)
	if err != nil {
		return err
	}

	// also dump pod logs for custom label selectors.
	for _, include := range includes {
		podLabelSelector := metav1.ListOptions{
			LabelSelector: include,
		}

		err = dumpPods(ctx, c, namespace, out, podLabelSelector, logLines)
		if err != nil {
			return err
		}
	}

	return nil
}

func dumpAll(ctx context.Context, c client.Client, namespace string, out io.Writer, logLines int, includes []string) error {
	yaksClient, err := versioned.NewForConfig(c.GetConfig())
	if err != nil {
		return err
	}

	appLabelSelector := metav1.ListOptions{
		LabelSelector: config.DefaultAppLabel,
	}

	err = dumpOperator(ctx, c, yaksClient, namespace, out, logLines)
	if err != nil {
		return err
	}

	tests, err := yaksClient.YaksV1alpha1().Tests(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "Found %d test(s):\n", len(tests.Items))
	for i := range tests.Items {
		err = printObject(&tests.Items[i], out)
		if err != nil {
			return err
		}
	}

	cms, err := c.CoreV1().ConfigMaps(namespace).List(ctx, appLabelSelector)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "Found %d configmap(s):\n", len(cms.Items))
	for i := range cms.Items {
		err = printObject(&cms.Items[i], out)
		if err != nil {
			return err
		}
	}

	deployments, err := c.AppsV1().Deployments(namespace).List(ctx, appLabelSelector)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "Found %d deployment(s):\n", len(deployments.Items))
	for i := range deployments.Items {
		err = printObject(&deployments.Items[i], out)
		if err != nil {
			return err
		}
	}

	jobs, err := c.BatchV1().Jobs(namespace).List(ctx, appLabelSelector)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "Found %d job(s):\n", len(jobs.Items))
	for i := range jobs.Items {
		err = printObject(&jobs.Items[i], out)
		if err != nil {
			return err
		}
	}

	err = dumpPods(ctx, c, namespace, out, appLabelSelector, logLines)
	if err != nil {
		return err
	}

	// also dump pod logs for custom label selectors.
	for _, include := range includes {
		podLabelSelector := metav1.ListOptions{
			LabelSelector: include,
		}

		err = dumpPods(ctx, c, namespace, out, podLabelSelector, logLines)
		if err != nil {
			return err
		}
	}

	return nil
}

func dumpOperator(ctx context.Context, c client.Client, yaksClient *versioned.Clientset, namespace string, out io.Writer, logLines int) error {
	instances, err := yaksClient.YaksV1alpha1().Instances(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "Found %d operator instance(s):\n", len(instances.Items))
	for i := range instances.Items {
		err = printObject(&instances.Items[i], out)
		if err != nil {
			return err
		}

		operatorPodSelector := metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=operator", cfg.ComponentLabel),
		}

		err = dumpPods(ctx, c, namespace, out, operatorPodSelector, logLines)
		if err != nil {
			return err
		}
	}

	if len(instances.Items) == 0 {
		// looking for global operator instance
		instance, err := v1alpha1.FindGlobalInstance(ctx, c)
		if err != nil && !errors.IsForbidden(err) {
			return err
		}

		if instance != nil {
			fmt.Fprintf(out, "Found global operator instance:\n")
			err = printObject(instance, out)
			if err != nil {
				return err
			}

			operatorPodSelector := metav1.ListOptions{
				LabelSelector: fmt.Sprintf("%s=operator", cfg.ComponentLabel),
			}

			err = dumpPods(ctx, c, instance.Namespace, out, operatorPodSelector, logLines)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func dumpPods(ctx context.Context, c client.Client, namespace string, out io.Writer, selector metav1.ListOptions, logLines int) error {
	podList, err := c.CoreV1().Pods(namespace).List(ctx, selector)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "\nFound %d pod(s):\n", len(podList.Items))
	for _, pod := range podList.Items {
		fmt.Fprintf(out, "name=%s\n", pod.Name)
		dumpConditions("  ", pod.Status.Conditions, out)
		fmt.Fprintf(out, "  logs:\n")
		var allContainers []v1.Container
		allContainers = append(allContainers, pod.Spec.InitContainers...)
		allContainers = append(allContainers, pod.Spec.Containers...)
		for _, container := range allContainers {
			pad := "    "
			fmt.Fprintf(out, "%s%s\n", pad, container.Name)
			err := dumpLogs(ctx, c, fmt.Sprintf("%s> ", pad), namespace, pod.Name, container.Name, out, logLines)
			if err != nil {
				fmt.Fprintf(out, "%sERROR while reading the logs: %v\n", pad, err)
			}
		}
	}

	return nil
}

func dumpConditions(prefix string, conditions []v1.PodCondition, out io.Writer) {
	for _, cond := range conditions {
		fmt.Fprintf(out, "%scondition type=%s, status=%s, reason=%s, message=%q\n", prefix, cond.Type, cond.Status, cond.Reason, cond.Message)
	}
}

func dumpLogs(ctx context.Context, c client.Client, prefix string, namespace string, name string, container string, out io.Writer, logLines int) error {
	options := v1.PodLogOptions{
		Container: container,
	}

	if lines := int64(logLines); lines > 0 {
		options.TailLines = &lines
	}

	stream, err := c.CoreV1().Pods(namespace).GetLogs(name, &options).Stream(ctx)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(stream)
	printed := false
	for scanner.Scan() {
		printed = true
		fmt.Fprintf(out, "%s%s\n", prefix, scanner.Text())
	}
	if !printed {
		fmt.Fprintf(out, "%s[no logs available]\n", prefix)
	}
	return stream.Close()
}

func printObject(ref runtime.Object, out io.Writer) error {
	printData, err := kubernetes.ToYAML(ref)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "---\n%s\n---\n", string(printData))

	return nil
}
