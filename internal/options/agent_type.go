package options

import "errors"

type AgentType string

const (
	AllAgentType     AgentType = "all"
	ClusterAgentType AgentType = "cluster"
	NodeAgentType    AgentType = "node"
	LegacyAgentType  AgentType = "legacy"
)

var InvalidAgentTypeErr = errors.New("--agent can only be node, cluster, all or legacy")

func NewAgentType(value string) (AgentType, error) {
	switch value {
	case "all":
		return AllAgentType, nil
	case "cluster":
		return ClusterAgentType, nil
	case "node":
		return NodeAgentType, nil
	case "legacy":
		return LegacyAgentType, nil
	}
	return "", InvalidAgentTypeErr
}

func (a AgentType) String() string {
	return string(a)
}

func (a *AgentType) Set(value string) error {
	var err error
	*a, err = NewAgentType(value)
	return err
}

func (a AgentType) Type() string {
	return "string"
}

func (a AgentType) ScrapeCluster() bool {
	return a == AllAgentType || a == LegacyAgentType
}

func (a AgentType) ScrapeNodes() string {
	switch a {
	case AllAgentType:
		return "all"

	case LegacyAgentType:
		fallthrough
	case NodeAgentType:
		return "own"

	case ClusterAgentType:
		fallthrough
	default:
		return "none"
	}
}
