package discovery

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

func ResourceName(kind string, meta metav1.ObjectMeta) string {
	if kind == ServiceType.String() {
		return meta.Namespace + "-" + kind + "-" + meta.Name
	}
	return kind + "-" + meta.Name
}
