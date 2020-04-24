/*
Copyright 2020 The Kubernetes Authors.

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

package controllers

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/util/json"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	shipsv1beta1 "github.com/crdAPIDemo/api/v1beta1"
)

// SloopReconciler reconciles a Sloop object
type SloopReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=ships.k8s.io,resources=sloops,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ships.k8s.io,resources=sloops/status,verbs=get;update;patch

func (r *SloopReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("sloop", req.NamespacedName)

	// your logic here
	// instance 和 spec 配置
	instance := &shipsv1beta1.Sloop{}
	if err := r.Get(ctx, req.NamespacedName, instance); err != nil {
		log.Info("unable to fetch vm : %v", err)
	} else {
		r.Printlog(instance)
	}

	// deploy
	deploy := &appsv1.Deployment{}
	if err := r.Get(context.TODO(), req.NamespacedName, deploy); err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new Deploy")

		// 创建关联资源
		// 1. 创建 Deploy
		deploy := shipsv1beta1.NewDeploy(instance)
		if err := r.Create(context.TODO(), deploy); err != nil {
			return reconcile.Result{}, err
		}

		// 2. 创建 Service
		service := shipsv1beta1.NewService(instance)
		if err := r.Create(context.TODO(), service); err != nil {
			return reconcile.Result{}, err
		}

		// 3. 关联 Annotations
		data, _ := json.Marshal(instance.Spec)
		if instance.Annotations != nil {
			instance.Annotations["spec"] = string(data)
		} else {
			instance.Annotations = map[string]string{"spec": string(data)}
		}

		if err := r.Update(context.TODO(), instance); err != nil {
			return reconcile.Result{}, nil
		}

		log.Info("Create a new Deploy done")
		return reconcile.Result{}, nil
	}

	// 比较新旧对象
	oldspec := shipsv1beta1.SloopSpec{}
	if instance.Annotations == nil {
		instance.Annotations = make(map[string]string)
/*	}else {

		if err := json.Unmarshal([]byte(instance.Annotations["spec"]), oldspec); err != nil {
			return reconcile.Result{}, err
		}*/
	}

	if !reflect.DeepEqual(instance.Spec, oldspec) {
		log.Info("Updating the old Deploy")

		// 更新关联资源
		newDeploy := shipsv1beta1.NewDeploy(instance)
		oldDeploy := &appsv1.Deployment{}
		if err := r.Get(context.TODO(), req.NamespacedName, oldDeploy); err != nil {
			return reconcile.Result{}, err
		}

		oldDeploy.Spec = newDeploy.Spec
		if err := r.Update(context.TODO(), oldDeploy); err != nil {
			return reconcile.Result{}, err
		}

		newService := shipsv1beta1.NewService(instance)
		oldService := &corev1.Service{}
		if err := r.Get(context.TODO(), req.NamespacedName, oldService); err != nil {
			return reconcile.Result{}, err
		}

		// 根据错误提示，增加的 ClusterIP 处理
		clusterIP := oldService.Spec.ClusterIP
		oldService.Spec = newService.Spec // 示例代码只做了这一步
		oldService.Spec.ClusterIP = clusterIP

		if err := r.Update(context.TODO(), oldService); err != nil {
			return reconcile.Result{}, err
		}

		log.Info("Update a new Deploy done")
		return reconcile.Result{}, nil
	}


	return ctrl.Result{}, nil
}

func (r *SloopReconciler) Printlog(instance *shipsv1beta1.Sloop){
	fmt.Println("INFO:", instance.Spec.Cpu, instance.Spec.Memory)
}

func (r *SloopReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&shipsv1beta1.Sloop{}).
		Complete(r)
}
