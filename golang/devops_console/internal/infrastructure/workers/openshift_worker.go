package workers

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type OpenShiftWorker struct {
	clientset *kubernetes.Clientset
    namespace string
}


func NewOpenShiftWorker(kubeconfig, namespace string) (*OpenShiftWorker, error) {
    config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
    if err != nil {
        return nil, fmt.Errorf("failed to build config: %v", err)
    }

    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        return nil, fmt.Errorf("failed to create clientset: %v", err)
    }

    return &OpenShiftWorker{
        clientset: clientset,
        namespace: namespace,
    }, nil
}
func (w *OpenShiftWorker) LaunchJob(ctx context.Context, name string, config map[string]interface{}) (string, error) {
    job := &batchv1.Job{
        ObjectMeta: metav1.ObjectMeta{
            Name: name,
        },
        Spec: batchv1.JobSpec{
            Template: corev1.PodTemplateSpec{
                Spec: corev1.PodSpec{
                    Containers: []corev1.Container{
                        {
                            Name:    "job",
                            Image:   config["image"].(string),
                            Command: []string{"/bin/sh", "-c", config["command"].(string)},
                        },
                    },
                    RestartPolicy: corev1.RestartPolicyNever,
                },
            },
        },
    }

    createdJob, err := w.clientset.BatchV1().Jobs(w.namespace).Create(ctx, job, metav1.CreateOptions{})
    if err != nil {
        return "", fmt.Errorf("failed to create job: %v", err)
    }

    return createdJob.Name, nil
}

func (w *OpenShiftWorker) GetJobStatus(ctx context.Context, name string) (string, error) {
    job, err := w.clientset.BatchV1().Jobs(w.namespace).Get(ctx, name, metav1.GetOptions{})
    if err != nil {
        return "", fmt.Errorf("failed to get job status: %v", err)
    }

    if job.Status.Succeeded > 0 {
        return "Completed", nil
    } else if job.Status.Failed > 0 {
        return "Failed", nil
    } else if job.Status.Active > 0 {
        return "Running", nil
    }

    return "Pending", nil
}

func (w *OpenShiftWorker) MonitorJob(ctx context.Context, name string) (<-chan string, error) {
    statusChan := make(chan string)

    go func() {
        defer close(statusChan)

        for {
            select {
            case <-ctx.Done():
                return
            default:
                status, err := w.GetJobStatus(ctx, name)
                if err != nil {
                    statusChan <- fmt.Sprintf("Error: %v", err)
                    return
                }

                statusChan <- status

                if status == "Completed" || status == "Failed" {
                    return
                }

                time.Sleep(5 * time.Second)
            }
        }
    }()

    return statusChan, nil
}

func (w *OpenShiftWorker) StreamLogs(ctx context.Context, name string) (<-chan string, error) {
    logChan := make(chan string)

    go func() {
        defer close(logChan)

        for {
            pods, err := w.clientset.CoreV1().Pods(w.namespace).List(ctx, metav1.ListOptions{
                LabelSelector: fmt.Sprintf("job-name=%s", name),
            })
            if err != nil {
                logChan <- fmt.Sprintf("Error listing pods: %v", err)
                return
            }

            if len(pods.Items) == 0 {
                time.Sleep(1 * time.Second)
                continue
            }

            podName := pods.Items[0].Name
            req := w.clientset.CoreV1().Pods(w.namespace).GetLogs(podName, &corev1.PodLogOptions{
                Follow: true,
            })

            stream, err := req.Stream(ctx)
            if err != nil {
                logChan <- fmt.Sprintf("Error opening log stream: %v", err)
                return
            }
            defer stream.Close()

            reader := bufio.NewReader(stream)
            for {
                line, err := reader.ReadString('\n')
                if err != nil {
                    if err == io.EOF {
                        return
                    }
                    logChan <- fmt.Sprintf("Error reading log: %v", err)
                    return
                }

                select {
                case logChan <- line:
                case <-ctx.Done():
                    return
                }
            }
        }
    }()

    return logChan, nil
}
