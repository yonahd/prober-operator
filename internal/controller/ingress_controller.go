/*
Copyright 2024 yonahd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/yaml"
)

const (
	probeFinalizer = "prober.io/prober.finalizer"
)

var ProbeTemplate = monitoringv1.Probe{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "probeName",
		Namespace: "namespace",
		Labels:    map[string]string{"placeholderLabelKey": "placeholderLabelValue"},
	},
	Spec: monitoringv1.ProbeSpec{
		Interval: "30s",
		Module:   "http_2xx",
		ProberSpec: monitoringv1.ProberSpec{
			URL: "proberURL",
		},
		Targets: monitoringv1.ProbeTargets{
			StaticConfig: &monitoringv1.ProbeTargetStaticConfig{
				Targets: []string{"domain"},
			},
		},
	},
}

// IngressReconciler reconciles an Ingress object
type IngressReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses/finalizers,verbs=update
//+kubebuilder:rbac:groups=monitoring.coreos.com,resources=probes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=monitoring.coreos.com,resources=probes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=monitoring.coreos.com,resources=probes/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Ingress object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.0/pkg/reconcile
func (r *IngressReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	ingress := &networkingv1.Ingress{}
	if err := r.Get(ctx, req.NamespacedName, ingress); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger.Info("Ingress", "Name", ingress.Name, "Namespace", ingress.Namespace)

	// TODO handle error
	probeExists, _ := r.ProbeFinder(ctx, ingress.Name, ingress.Namespace)

	if ingress.GetLabels()["monitor"] == "true" {

		if !ingress.GetDeletionTimestamp().IsZero() {
			err := r.ProbeCleaner(ctx, *ingress)
			if err != nil {
				return ctrl.Result{}, err
			}
		} else {
			logger.Info("Found candidate", "Name", ingress.Name, "Namespace", ingress.Namespace)
			if probeExists == true {
				logger.Info("Probe exists for", "Name", ingress.Name, "Namespace", ingress.Namespace)
			} else {
				logger.Info("Creating probe for", "Name", ingress.Name, "Namespace", ingress.Namespace)
				err := r.createProbeForIngress(ctx, ingress.Name, ingress.Namespace, ingress.Spec.Rules[0].Host)
				if err != nil {
					return ctrl.Result{}, err
				}

				r.AddFinalizerIfNeeded(ingress)
				if err := r.Update(context.Background(), ingress); err != nil {
					logger.Error(err, "Failed to update Ingress with finalizer")
					return ctrl.Result{}, err
				}
			}
		}

	} else {
		if probeExists == true {
			err := r.ProbeCleaner(ctx, *ingress)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{}, nil
}

func (r *IngressReconciler) AddFinalizerIfNeeded(ingress *networkingv1.Ingress) {
	if !controllerutil.ContainsFinalizer(ingress, probeFinalizer) {
		ingress.SetFinalizers(append(ingress.GetFinalizers(), probeFinalizer))
	}
}

func (r *IngressReconciler) ProbeCleaner(ctx context.Context, ingress networkingv1.Ingress) error {
	logger := log.FromContext(ctx)

	logger.Info("Deleting probe for", "Name", ingress.Name, "Namespace", ingress.Namespace)
	err := r.deleteProbeForIngress(ctx, ingress.Name, ingress.Namespace)
	if err != nil {
		return err
	}

	controllerutil.RemoveFinalizer(&ingress, probeFinalizer)
	if err := r.Update(context.Background(), &ingress); err != nil {
		logger.Error(err, "Failed to update Ingress without finalizer")
		return err
	}
	return nil
}

func (r *IngressReconciler) ProbeFinder(ctx context.Context, ingressName, namespace string) (bool, error) {
	probeName := fmt.Sprintf("%s-probe", ingressName)

	probe := &monitoringv1.Probe{}

	err := r.Get(ctx, client.ObjectKey{Namespace: namespace, Name: probeName}, probe)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return false, err // Other errors
		}

		return false, nil // PrometheusProbe not found
	}

	return true, nil // PrometheusProbe found
}

func (r *IngressReconciler) createProbeForIngress(ctx context.Context, ingressName, namespace, domain string) error {

	probeName := fmt.Sprintf("%s-probe", ingressName)

	probe := ProbeTemplate.DeepCopy()

	labels, proberURL, err := r.getBlackboxInfo(ctx)
	if err != nil {
		return fmt.Errorf("error getting probe info: %v", err)
	}
	probe.ObjectMeta.Name = probeName
	probe.ObjectMeta.Namespace = namespace
	probe.ObjectMeta.Labels = labels
	probe.Spec.ProberSpec.URL = proberURL
	probe.Spec.Targets.StaticConfig.Targets = []string{fmt.Sprintf("https://%s", domain)}

	err = r.Create(ctx, probe)
	if err != nil {
		return fmt.Errorf("error creating probe: %v", err)
	}

	fmt.Printf("Created probe for Ingress %s\n", ingressName)
	return nil
}

func (r *IngressReconciler) deleteProbeForIngress(ctx context.Context, ingressName, namespace string) error {
	probeName := fmt.Sprintf("%s-probe", ingressName)
	probe := &monitoringv1.Probe{}

	err := r.Get(ctx, client.ObjectKey{Namespace: namespace, Name: probeName}, probe)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return err
		}
	}
	err = r.Delete(ctx, probe)
	if err != nil {
		return err
	}

	return nil
}

func (r *IngressReconciler) getBlackboxInfo(ctx context.Context) (map[string]string, string, error) {
	configMapName := "prober-blackbox-config"

	configMapList := &corev1.ConfigMapList{}
	err := r.List(ctx, configMapList, client.InNamespace(""))
	if err != nil {
		return nil, "", err
	}

	var configMap *corev1.ConfigMap
	for _, cm := range configMapList.Items {
		if cm.Name == configMapName {
			configMap = &cm
			break
		}
	}

	probeUrl := "prober-operator-prometheus-blackbox-exporter:9115"

	if configMapData, ok := configMap.Data["proberURL"]; ok {
		probeUrl = configMapData
	}

	labels := make(map[string]string)
	labelsData, ok := configMap.Data["labels"]
	if ok {
		err = yaml.Unmarshal([]byte(labelsData), &labels)
		if err != nil {
			return nil, "", err
		}
	} else {
		labels = map[string]string{
			"app.kubernetes.io/instance": "kube-prometheus-stack",
			"release":                    "kube-prometheus-stack",
		}
	}

	return labels, probeUrl, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *IngressReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		// Uncomment the following line adding a pointer to an instance of the controlled resource as an argument
		For(&networkingv1.Ingress{}).
		Complete(r)
}
