// kubectl-why — explain WHY a pod isn't running, in one clear answer.
//
//	Usage:  kubectl why <pod> [-n namespace]
package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	dim    = "\033[2m"
	reset  = "\033[0m"
)

func main() {
	pod, namespace, nsExplicit := parseArgs(os.Args[1:])
	if pod == "" {
		fmt.Fprintln(os.Stderr, "Usage: kubectl why <pod> [-n namespace]")
		os.Exit(1)
	}

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules, &clientcmd.ConfigOverrides{})

	if !nsExplicit {
		if ns, _, err := kubeConfig.Namespace(); err == nil && ns != "" {
			namespace = ns
		} else {
			namespace = "default"
		}
	}

	restConfig, err := kubeConfig.ClientConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s❌ Can't load kubeconfig: %v%s\n", red, err, reset)
		os.Exit(1)
	}
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s❌ Can't create Kubernetes client: %v%s\n", red, err, reset)
		os.Exit(1)
	}

	ctx := context.Background()
	p, err := clientset.CoreV1().Pods(namespace).Get(ctx, pod, metav1.GetOptions{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s❌ Pod %q not found in namespace %q.%s\n", red, pod, namespace, reset)
		os.Exit(1)
	}

	diagnose(ctx, clientset, p, namespace)
}

func parseArgs(args []string) (pod, namespace string, nsExplicit bool) {
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-n", "--namespace":
			if i+1 < len(args) {
				namespace, nsExplicit = args[i+1], true
				i++
			}
		case "-h", "--help":
			fmt.Println("Usage: kubectl why <pod> [-n namespace]")
			os.Exit(0)
		default:
			pod = args[i]
		}
	}
	return
}

func diagnose(ctx context.Context, cs *kubernetes.Clientset, p *corev1.Pod, ns string) {
	phase := string(p.Status.Phase)

	allReady := len(p.Status.ContainerStatuses) > 0
	for _, c := range p.Status.ContainerStatuses {
		if !c.Ready {
			allReady = false
		}
	}
	if phase == "Running" && allReady {
		fmt.Printf("%s✅ Pod %q is Running and all containers are ready.%s\n", green, p.Name, reset)
		return
	}

	fmt.Printf("%s🔍 Diagnosing pod %q (phase: %s)…%s\n\n", yellow, p.Name, phase, reset)

	// 1) containers stuck in "waiting"
	waiting := false
	for _, c := range p.Status.ContainerStatuses {
		if c.State.Waiting == nil {
			continue
		}
		waiting = true
		reason, msg := c.State.Waiting.Reason, c.State.Waiting.Message
		switch reason {
		case "CrashLoopBackOff":
			fmt.Printf("%s❌ Container %q: CrashLoopBackOff%s\n", red, c.Name, reset)
			fmt.Println("   └─ It starts, crashes, and keeps restarting.")
			fmt.Printf("   └─ Logs:  kubectl logs %s -n %s -c %s --previous\n", p.Name, ns, c.Name)
		case "ImagePullBackOff", "ErrImagePull":
			fmt.Printf("%s❌ Container %q: %s%s\n", red, c.Name, reason, reset)
			fmt.Println("   └─ Kubernetes can't pull the image.")
			fmt.Println("   └─ Check: image name/tag typo, or private registry needs an imagePullSecret.")
		case "CreateContainerConfigError":
			fmt.Printf("%s❌ Container %q: CreateContainerConfigError%s\n", red, c.Name, reset)
			fmt.Println("   └─ Usually a missing ConfigMap or Secret referenced in env/volumes.")
			if msg != "" {
				fmt.Printf("   └─ %s\n", truncate(msg, 160))
			}
		default:
			fmt.Printf("%s❌ Container %q: %s%s\n", red, c.Name, reason, reset)
			if msg != "" {
				fmt.Printf("   └─ %s\n", truncate(msg, 160))
			}
		}
		fmt.Println()
	}
	if waiting {
		printEvents(ctx, cs, p, ns)
		return
	}

	// 2) Pending — a scheduling problem
	if phase == "Pending" {
		var schedMsg string
		for _, cond := range p.Status.Conditions {
			if cond.Type == corev1.PodScheduled && cond.Status == corev1.ConditionFalse {
				schedMsg = cond.Message
			}
		}
		fmt.Printf("%s❌ Pod %q is Pending — not scheduled to any node.%s\n", red, p.Name, reset)
		if schedMsg != "" {
			fmt.Printf("   └─ Scheduler says: %s\n", schedMsg)
		}
		fmt.Println("   └─ Common causes: insufficient CPU/memory, node taints, unbound PVC.")
		fmt.Println()
		printEvents(ctx, cs, p, ns)
		return
	}

	// 3) a container terminated (e.g. OOMKilled)
	terminated := false
	for _, c := range p.Status.ContainerStatuses {
		if c.State.Terminated == nil {
			continue
		}
		terminated = true
		t := c.State.Terminated
		fmt.Printf("%s❌ Container %q terminated: %s (exit code %d)%s\n", red, c.Name, t.Reason, t.ExitCode, reset)
		if t.ExitCode == 137 {
			fmt.Println("   └─ Exit 137 = OOMKilled — it ran out of memory (raise limits.memory).")
		}
		fmt.Printf("   └─ Logs:  kubectl logs %s -n %s -c %s --previous\n", p.Name, ns, c.Name)
		fmt.Println()
	}
	if terminated {
		printEvents(ctx, cs, p, ns)
		return
	}

	fmt.Printf("%s⚠️  Pod is unhealthy but no known pattern matched. Showing events:%s\n", yellow, reset)
	printEvents(ctx, cs, p, ns)
}

func printEvents(ctx context.Context, cs *kubernetes.Clientset, p *corev1.Pod, ns string) {
	fmt.Printf("%sRecent events:%s\n", dim, reset)
	events, err := cs.CoreV1().Events(ns).List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s", p.Name),
	})
	if err != nil || len(events.Items) == 0 {
		fmt.Println("   (no recent events)")
		return
	}
	sort.Slice(events.Items, func(i, j int) bool {
		return events.Items[i].LastTimestamp.Before(&events.Items[j].LastTimestamp)
	})
	items := events.Items
	if len(items) > 5 {
		items = items[len(items)-5:]
	}
	for _, e := range items {
		fmt.Printf("   %-8s %-10s %s\n", e.Type, e.Reason, truncate(strings.TrimSpace(e.Message), 90))
	}
}

func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) > n {
		return s[:n] + "…"
	}
	return s
}
