package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/arangodb/go-driver"
	arango "github.com/inksnw/gorm-arango"
	"gorm.io/gorm"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

const GraphName = "MizarGraph"

func CreateGraph(db *gorm.DB, cfg *Config) error {
	ctx := context.TODO()
	dialector, _ := db.Config.Dialector.(arango.Dialector)
	database, err := dialector.Client.Database(ctx, cfg.Database)
	if err != nil {
		return err
	}
	exists, err := database.GraphExists(ctx, GraphName)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	edgeDefinition := driver.EdgeDefinition{
		Collection: "edges",
		From:       []string{"resources"},
		To:         []string{"resources"},
	}
	options := &driver.CreateGraphOptions{
		EdgeDefinitions: []driver.EdgeDefinition{edgeDefinition},
	}
	graph, err := database.CreateGraphV2(ctx, GraphName, options)
	msg := fmt.Sprintf("graph %s create successed", graph.Name())
	db.Logger.Info(ctx, msg)
	return err
}

func (s *ResourceStorage) addEdge(obj runtime.Object, resource Resource) (err error) {
	switch resource.Kind {
	case "PersistentVolume":
		err = s.PersistentVolume(obj, resource)
	case "Pod":
		err = s.Pod(obj, resource)
	case "ReplicaSet":
		err = s.ReplicaSet(obj, resource)
	case "Deployment":
		err = s.Deployment(obj, resource)
	case "Service":
		err = s.Service(obj, resource)
	case "StatefulSet":
		err = s.StatefulSet(obj, resource)
	case "DaemonSet":
		err = s.DaemonSet(obj, resource)
	case "ConfigMap":
		err = s.ConfigMap(obj, resource)
	case "Secret":
		err = s.Secret(obj, resource)
	case "Ingress":
		err = s.Ingress(obj, resource)
	case "PersistentVolumeClaim":
		err = s.PersistentVolumeClaim(obj, resource)
	}
	return err
}

func (s *ResourceStorage) Pod(obj runtime.Object, resource Resource) error {
	metaInfo, _ := meta.Accessor(obj)
	own := metaInfo.GetOwnerReferences()
	if len(own) == 0 {
		return nil
	}
	query := s.db.WithContext(context.TODO()).Model(&Resource{}).Select("object")
	var resources []json.RawMessage

	key := fmt.Sprintf("%s", own[0].UID)
	where := map[string]any{
		"cluster":   resource.Cluster,
		"kind":      "ReplicaSet",
		"namespace": resource.Namespace,
		"_key":      key,
	}
	query.Where(where).Find(&resources)
	for _, i := range resources {
		unst1, err := Decode(i)
		for _, ownRS := range unst1.GetOwnerReferences() {
			err = s.createEdge(ownRS.UID, metaInfo.GetUID())
			if err != nil {
				return err
			}
		}
	}
	return nil
}
func (s *ResourceStorage) ReplicaSet(obj runtime.Object, resource Resource) error {
	metaInfo, _ := meta.Accessor(obj)
	own := metaInfo.GetOwnerReferences()
	if len(own) > 0 {
		from := own[0].UID
		to := metaInfo.GetUID()
		err := s.createEdge(from, to)
		return err
	}
	return nil
}

func (s *ResourceStorage) PersistentVolume(obj runtime.Object, resource Resource) error {
	metaInfo, _ := meta.Accessor(obj)
	unst, _ := obj.(*unstructured.Unstructured)

	pv := &corev1.PersistentVolume{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(unst.UnstructuredContent(), pv)
	if err != nil {
		return err
	}

	if pvcRef := pv.Spec.ClaimRef; pvcRef != nil {
		from := pvcRef.UID
		to := metaInfo.GetUID()
		err = s.createEdge(from, to)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *ResourceStorage) Deployment(obj runtime.Object, resource Resource) error {
	metaInfo, _ := meta.Accessor(obj)
	return s.withLabel(resource, metaInfo, "Pod")
}

func (s *ResourceStorage) Service(obj runtime.Object, resource Resource) error {
	unst, _ := obj.(*unstructured.Unstructured)
	svc := &corev1.Service{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(unst.UnstructuredContent(), svc)
	if err != nil {
		return err
	}

	query := s.db.WithContext(context.TODO()).Model(&Resource{}).Select("object")
	var resources []json.RawMessage
	where := map[string]any{"cluster": resource.Cluster, "kind": "Pod", "namespace": resource.Namespace}
	query = queryLabel(query, svc.Spec.Selector)
	query.Where(where).Find(&resources)
	for _, i := range resources {
		unst1, err := Decode(i)
		err = s.createEdge(unst.GetUID(), unst1.GetUID())
		if err != nil {
			return err
		}
	}
	return nil
}
func (s *ResourceStorage) StatefulSet(obj runtime.Object, resource Resource) error {
	metaInfo, _ := meta.Accessor(obj)

	return s.withLabel(resource, metaInfo, "Pod")
}

func (s *ResourceStorage) withLabel(resource Resource, metaInfo metav1.Object, kind string) error {
	query := s.db.WithContext(context.TODO()).Model(&Resource{}).Select("object")
	var resources []json.RawMessage
	where := map[string]any{"cluster": resource.Cluster, "kind": kind, "namespace": resource.Namespace}
	query = queryLabel(query, metaInfo.GetLabels())
	query.Where(where).Find(&resources)
	for _, i := range resources {
		unst, err := Decode(i)
		if err != nil {
			return err
		}
		err = s.createEdge(metaInfo.GetUID(), unst.GetUID())
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *ResourceStorage) DaemonSet(obj runtime.Object, resource Resource) error {
	metaInfo, _ := meta.Accessor(obj)
	return s.withLabel(resource, metaInfo, "Pod")
}

func (s *ResourceStorage) ConfigMap(obj runtime.Object, resource Resource) error {
	metaInfo, _ := meta.Accessor(obj)
	query := s.db.WithContext(context.TODO()).Model(&Resource{}).Select("object")
	var resources []json.RawMessage
	where := map[string]any{"cluster": resource.Cluster, "kind": "Pod", "namespace": resource.Namespace}
	query.Where(where).Find(&resources)
	for _, i := range resources {
		unst, err := Decode(i)
		if err != nil {
			return err
		}

		pod := &corev1.Pod{}
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(unst.UnstructuredContent(), pod)
		if err != nil {
			return err
		}

		for _, volume := range pod.Spec.Volumes {
			if volume.ConfigMap != nil && volume.ConfigMap.Name == metaInfo.GetName() {
				err = s.createEdge(metaInfo.GetUID(), pod.UID)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (s *ResourceStorage) Secret(obj runtime.Object, resource Resource) error {
	metaInfo, _ := meta.Accessor(obj)
	query := s.db.WithContext(context.TODO()).Model(&Resource{}).Select("object")
	var resources []json.RawMessage
	where := map[string]any{"cluster": resource.Cluster, "kind": "Pod", "namespace": resource.Namespace}
	rv := query.Where(where).Find(&resources)
	if rv.Error != nil {
		return rv.Error
	}
	for _, i := range resources {
		unst, err := Decode(i)
		if err != nil {
			return err
		}
		pod := &corev1.Pod{}
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(unst.UnstructuredContent(), pod)
		if err != nil {
			return err
		}

		for _, volume := range pod.Spec.Volumes {
			if volume.Secret != nil && volume.Secret.SecretName == metaInfo.GetName() {
				err = s.createEdge(metaInfo.GetUID(), pod.UID)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (s *ResourceStorage) Ingress(obj runtime.Object, resource Resource) error {
	unst, _ := obj.(*unstructured.Unstructured)

	ingress := &networkingv1.Ingress{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(unst.UnstructuredContent(), ingress)
	if err != nil {
		return err
	}

	for _, rule := range ingress.Spec.Rules {
		for _, path := range rule.HTTP.Paths {
			serviceName := path.Backend.Service.Name
			query := s.db.WithContext(context.TODO()).Model(&Resource{}).Select("object")
			var resources []json.RawMessage
			where := map[string]any{
				"cluster":   resource.Cluster,
				"kind":      "Service",
				"namespace": resource.Namespace,
				"name":      serviceName}
			query.Where(where).Find(&resources)
			for _, i := range resources {
				unst1, err := Decode(i)
				if err != nil {
					return err
				}
				err = s.createEdge(unst.GetUID(), unst1.GetUID())
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (s *ResourceStorage) PersistentVolumeClaim(obj runtime.Object, resource Resource) error {
	unst, _ := obj.(*unstructured.Unstructured)

	pvc := &corev1.PersistentVolumeClaim{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(unst.UnstructuredContent(), pvc)
	if err != nil {
		return err
	}

	query := s.db.WithContext(context.TODO()).Model(&Resource{}).Select("object")
	var resources []json.RawMessage
	where := map[string]any{
		"cluster":   resource.Cluster,
		"kind":      "PersistentVolume",
		"namespace": resource.Namespace,
		"name":      pvc.Spec.VolumeName}
	query.Where(where).Find(&resources)
	for _, i := range resources {
		unst1, err := Decode(i)
		if err != nil {
			return err
		}
		err = s.createEdge(unst.GetUID(), unst1.GetUID())
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *ResourceStorage) createEdge(from, to types.UID) error {
	fromStr, toStr := string(from), string(to)
	GraphKey := fmt.Sprintf("%s:%s", fromStr, toStr)
	GraphID := fmt.Sprintf("%s/%s", "edges", GraphKey)
	fromStr = fmt.Sprintf("%s/%s", "resources", fromStr)
	toStr = fmt.Sprintf("%s/%s", "resources", toStr)
	relationship := Edge{From: fromStr, To: toStr, GraphKey: GraphKey, GraphID: GraphID}
	err := s.db.Create(relationship).Error

	if arangoErr, ok := err.(driver.ArangoError); ok {
		if arangoErr.ErrorNum == driver.ErrArangoUniqueConstraintViolated {
			return nil
		}
	}
	return err
}

func Decode(data []byte) (unst *unstructured.Unstructured, err error) {
	decoder := unstructured.UnstructuredJSONScheme
	obj, _, err := decoder.Decode(data, nil, nil)
	unst, _ = obj.(*unstructured.Unstructured)
	return unst, err
}

func (s *ResourceStorage) deleteEdge() error {
	return nil
}

func queryLabel(query *gorm.DB, labelMap map[string]string) *gorm.DB {
	for k, v := range labelMap {
		jsonQuery := JSONQuery("object", "metadata", "labels", k)
		jsonQuery.Equal(v)
		query = query.Where(jsonQuery)
	}
	return query
}
