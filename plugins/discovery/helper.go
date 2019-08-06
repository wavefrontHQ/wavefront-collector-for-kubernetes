package discovery

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func EncodeMeta(tags map[string]string, kind string, meta metav1.ObjectMeta) {
	tags["kind"] = meta.Name
	if meta.Namespace != "" {
		tags["namespace"] = meta.Namespace
	}
}
