package flags

import (
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/httputil"

	"github.com/golang/glog"
)

// Decodes tags of the form "tag=key:value"
func DecodeTags(vals map[string][]string) map[string]string {
	if vals == nil {
		return nil
	}
	var tags map[string]string
	if len(vals["tag"]) > 0 {
		tags = make(map[string]string)
		tagList := vals["tag"]
		for _, tag := range tagList {
			s := strings.Split(tag, ":")
			if len(s) == 2 {
				k, v := s[0], s[1]
				tags[k] = v
			} else {
				glog.Warning("invalid tag ", tag)
			}
		}
	}
	return tags
}

func DecodeValue(vals map[string][]string, name string) string {
	value := ""
	if len(vals[name]) > 0 {
		value = vals[name][0]
	}
	return value
}

func DecodeDefaultValue(vals map[string][]string, name, defaultValue string) string {
	value := DecodeValue(vals, name)
	if value == "" {
		return defaultValue
	}
	return value
}

func DecodeBoolean(vals map[string][]string, name string) bool {
	value := false
	if len(vals[name]) > 0 {
		var err error
		value, err = strconv.ParseBool(vals[name][0])
		if err != nil {
			return false
		}
	}
	return value
}

func DecodeHTTPConfig(vals map[string][]string) httputil.ClientConfig {
	return httputil.ClientConfig{
		BearerToken:     DecodeValue(vals, "bearerToken"),
		BearerTokenFile: DecodeValue(vals, "bearerTokenFile"),
		TLSConfig: httputil.TLSConfig{
			CAFile:             DecodeValue(vals, "tlsCAFile"),
			CertFile:           DecodeValue(vals, "tlsCertFile"),
			KeyFile:            DecodeValue(vals, "tlsKeyFile"),
			ServerName:         DecodeValue(vals, "tlsServerName"),
			InsecureSkipVerify: DecodeBoolean(vals, "tlsInsecure"),
		},
	}
}

func ParseDuration(vals url.Values, prop string, def time.Duration) time.Duration {
	if len(vals[prop]) > 0 {
		res, err := time.ParseDuration(vals[prop][0])
		if err != nil {
			glog.Errorf("error parsing '%s' propertie: %v", prop, err)
		} else {
			return res
		}
	}
	return def
}
