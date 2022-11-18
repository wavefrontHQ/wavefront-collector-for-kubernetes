package wavefront

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCleanTags(t *testing.T) {
	t.Run("excludes tags in the exclude tag list", func(t *testing.T) {
		for _, excludedTagName := range excludeTagList {
			actual := map[string]string{excludedTagName: "some-value"}
			cleanTags(actual, maxWavefrontTags)
			assert.Equal(t, map[string]string{}, actual)
		}
	})

	t.Run("excludes tags with given prefixes", func(t *testing.T) {
		for _, excludedTagName := range excludeTagPrefixes {
			actual := map[string]string{excludedTagName + "/something": "some-value"}
			cleanTags(actual, maxWavefrontTags)
			assert.Equal(t, map[string]string{}, actual)
		}
	})

	t.Run("excludes empty tags", func(t *testing.T) {
		actual := map[string]string{"good-tag": ""}
		cleanTags(actual, maxWavefrontTags)
		assert.Equal(t, map[string]string{}, actual)
	})

	t.Run("de-duplicates tag values >= min dedupe value length characters when over capacity", func(t *testing.T) {
		tagGreaterThanMinLen := "some.hostname"
		assert.True(t, len(tagGreaterThanMinLen) >= minDedupeTagValueLen)

		tagEqualMinLen := "host1"
		assert.True(t, len(tagEqualMinLen) == minDedupeTagValueLen)

		tagLessThanMinLen := "host"
		assert.True(t, len(tagLessThanMinLen) < minDedupeTagValueLen)

		t.Run("when the tag names are different lengths", func(t *testing.T) {
			actual := map[string]string{"long-tag-name": tagGreaterThanMinLen, "shrt-tg": tagGreaterThanMinLen}
			cleanTags(actual, 1)
			assert.Equal(t, map[string]string{"shrt-tg": tagGreaterThanMinLen}, actual)
		})

		t.Run("when the tag names of the same length", func(t *testing.T) {
			actual := map[string]string{"dup2": tagGreaterThanMinLen, "dup1": tagGreaterThanMinLen}
			cleanTags(actual, 1)
			assert.Equal(t, map[string]string{"dup1": tagGreaterThanMinLen}, actual)
		})

		t.Run("when the duplicated values are < min len characters", func(t *testing.T) {
			actual := map[string]string{"a-tag": tagLessThanMinLen, "b-tag": tagLessThanMinLen}
			cleanTags(actual, 1)
			assert.Equal(t, map[string]string{"a-tag": tagLessThanMinLen, "b-tag": tagLessThanMinLen}, actual)
		})

		t.Run("when the duplicated values are equal min len characters", func(t *testing.T) {
			actual := map[string]string{"a-tag": tagEqualMinLen, "b-tag": tagEqualMinLen}
			cleanTags(actual, 1)
			assert.Equal(t, map[string]string{"a-tag": tagEqualMinLen}, actual)
		})

		t.Run("when under the max capacity", func(t *testing.T) {
			actual := map[string]string{"a-tag": tagLessThanMinLen, "b-tag": tagLessThanMinLen}
			cleanTags(actual, 2)
			assert.Equal(t, map[string]string{"a-tag": tagLessThanMinLen, "b-tag": tagLessThanMinLen}, actual)
		})
	})

	t.Run("limits example IaaS node info metric tags to max capacity ", func(t *testing.T) {
		t.Run("GKE example", func(t *testing.T) {
			actual := map[string]string{
				"cluster":                                "mamichael-gke--helm-21114",
				"nodename":                               "gke-mamichael-cluster-5-default-pool-5592f664-mkrr",
				"node_role":                              "worker",
				"os_image":                               "Container-Optimized OS from Google",
				"kubelet_version":                        "v1.23.8-gke.1900",
				"pod_cidr":                               "10.96.2.0/24",
				"internal_ip":                            "10.40.56.17",
				"kernel_version":                         "5.10.127+",
				"provider_id":                            "gce://wavefront-gcp-dev/us-central1-c/gke-mamichael-cluster-5-default-pool-5592f664-mkrr",
				"label.beta.kubernetes.io/arch":          "amd64",
				"label.beta.kubernetes.io/instance-type": "e2-standard-2",
				"label.beta.kubernetes.io/os":            "linux",
				"label.cloud.google.com/gke-boot-disk":   "pd-standard",
				"label.cloud.google.com/gke-container-runtime":   "containerd",
				"label.cloud.google.com/gke-cpu-scaling-level":   "2",
				"label.cloud.google.com/gke-max-pods-per-node":   "110",
				"label.cloud.google.com/gke-nodepool":            "default-pool",
				"label.cloud.google.com/gke-os-distribution":     "cos",
				"label.cloud.google.com/machine-family":          "e2",
				"label.failure-domain.beta.kubernetes.io/region": "us-central1",
				"label.failure-domain.beta.kubernetes.io/zone":   "us-central1-c",
				"label.kubernetes.io/arch":                       "amd64",
				"label.kubernetes.io/hostname":                   "gke-mamichael-cluster-5-default-pool-5592f664-3op5",
				"label.kubernetes.io/os":                         "linux",
				"label.node.kubernetes.io/instance-type":         "e2-standard-2",
				"label.topology.gke.io/zone":                     "us-central1-c",
				"label.topology.kubernetes.io/region":            "us-central1",
				"label.topology.kubernetes.io/zone":              "us-central1-c"}

			expectedCleanedTags := map[string]string{
				"cluster":         "mamichael-gke--helm-21114",
				"nodename":        "gke-mamichael-cluster-5-default-pool-5592f664-mkrr",
				"node_role":       "worker",
				"os_image":        "Container-Optimized OS from Google",
				"kubelet_version": "v1.23.8-gke.1900",
				"pod_cidr":        "10.96.2.0/24",
				"internal_ip":     "10.40.56.17",
				"kernel_version":  "5.10.127+",
				"provider_id":     "gce://wavefront-gcp-dev/us-central1-c/gke-mamichael-cluster-5-default-pool-5592f664-mkrr",
				"label.cloud.google.com/gke-max-pods-per-node": "110",
				"label.cloud.google.com/gke-nodepool":          "default-pool",
				"label.cloud.google.com/gke-os-distribution":   "cos",
				"label.cloud.google.com/machine-family":        "e2",
				"label.kubernetes.io/arch":                     "amd64",
				"label.kubernetes.io/hostname":                 "gke-mamichael-cluster-5-default-pool-5592f664-3op5",
				"label.kubernetes.io/os":                       "linux",
				"label.node.kubernetes.io/instance-type":       "e2-standard-2",
				"label.topology.gke.io/zone":                   "us-central1-c",
				"label.topology.kubernetes.io/region":          "us-central1"}
			cleanTags(actual, maxWavefrontTags)
			assert.Equal(t, maxWavefrontTags, len(actual))
			assert.Equal(t, expectedCleanedTags, actual)
		})
		t.Run("AKS example", func(t *testing.T) {
			actual := map[string]string{
				"cluster":                                "mamichael-aks-221116",
				"node_role":                              "worker",
				"nodename":                               "aks-agentpool-33708643-vmss000000",
				"type":                                   "node",
				"pod_cidr":                               "10.96.2.0/24",
				"internal_ip":                            "10.40.56.17",
				"kernel_version":                         "5.10.127+",
				"label.agentpool":                        "agentpool",
				"label.beta.kubernetes.io/arch":          "amd64",
				"label.beta.kubernetes.io/instance-type": "Standard_B4ms",
				"label.beta.kubernetes.io/os":            "linux",
				"label.failure-domain.beta.kubernetes.io/region":        "eastus",
				"label.failure-domain.beta.kubernetes.io/zone":          "0",
				"label.kubernetes.azure.com/agentpool":                  "agentpool",
				"label.kubernetes.azure.com/cluster":                    "MC_K8sSaaS_k8po-ci_eastus",
				"label.kubernetes.azure.com/kubelet-identity-client-id": "50903262-46af-4c4d-b1c7-440985c16284",
				"label.kubernetes.azure.com/mode":                       "system",
				"label.kubernetes.azure.com/node-image-version":         "AKSUbuntu-1804gen2containerd-2022.08.10",
				"label.kubernetes.azure.com/os-sku":                     "Ubuntu",
				"label.kubernetes.azure.com/role":                       "agent",
				"label.kubernetes.azure.com/storageprofile":             "managed",
				"label.kubernetes.azure.com/storagetier":                "Premium_LRS",
				"label.kubernetes.io/arch":                              "amd64",
				"label.kubernetes.io/hostname":                          "aks-agentpool-33708643-vmss000000",
				"label.kubernetes.io/os":                                "linux",
				"label.kubernetes.io/role":                              "agent",
				"label.node-role.kubernetes.io/agent":                   "",
				"label.node.kubernetes.io/instance-type":                "Standard_B4ms",
				"label.storageprofile":                                  "managed",
				"label.storagetier":                                     "Premium_LRS",
				"resource_id":                                           "/",
				"label.topology.disk.csi.azure.com/zone":                "",
				"label.topology.kubernetes.io/region":                   "eastus",
				"label.topology.kubernetes.io/zone":                     "0",
			}

			expected := map[string]string{
				"cluster":         "mamichael-aks-221116",
				"node_role":       "worker",
				"nodename":        "aks-agentpool-33708643-vmss000000",
				"type":            "node",
				"pod_cidr":        "10.96.2.0/24",
				"internal_ip":     "10.40.56.17",
				"kernel_version":  "5.10.127+",
				"label.agentpool": "agentpool",
				"label.failure-domain.beta.kubernetes.io/zone":  "0",
				"label.kubernetes.azure.com/node-image-version": "AKSUbuntu-1804gen2containerd-2022.08.10",
				"label.kubernetes.azure.com/os-sku":             "Ubuntu",
				"label.kubernetes.io/arch":                      "amd64",
				"label.kubernetes.io/os":                        "linux",
				"label.kubernetes.io/role":                      "agent",
				"label.node.kubernetes.io/instance-type":        "Standard_B4ms",
				"label.storageprofile":                          "managed",
				"label.storagetier":                             "Premium_LRS",
				"label.topology.kubernetes.io/region":           "eastus",
				"label.topology.kubernetes.io/zone":             "0",
			}

			cleanTags(actual, maxWavefrontTags)
			assert.Equal(t, maxWavefrontTags, len(actual))
			assert.Equal(t, expected, actual)
		})
	})
}

func TestIsAnEmptyTag(t *testing.T) {
	assert.True(t, isAnEmptyTag(""))
	assert.True(t, isAnEmptyTag("/"))
	assert.True(t, isAnEmptyTag("-"))
}
