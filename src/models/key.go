package models

import (
	"fmt"
	"net/http"
	"strings"
)

type (
	ProxieInfo struct {
		Id   ProxieKey
		Conf *Conf
	}
	ProxieTesterType string
	ProxieTester     interface {
		Name() ProxieTesterType
		Cron(*Conf) string
		RunTest(proxy *ProxieInfo, transport http.RoundTripper) (map[string]any, error)
	}
)

const CronJobKeyPrefix = "CronJob/"

type CronJobKey struct {
	ConfId string
	Type   ProxieTesterType
}

func (c *CronJobKey) ToKey() []byte {
	return []byte(CronJobKeyPrefix + strings.Join([]string{c.ConfId, string(c.Type)}, "::"))
}

func (c *CronJobKey) ToProxiePrefixKey() []byte {
	return []byte(ProxieKeyPrefix + c.ConfId)
}

func (c *CronJobKey) FromKey(_key []byte) error {
	key := strings.TrimPrefix(string(_key), CronJobKeyPrefix)
	parts := strings.Split(key, "::")
	if len(parts) != 2 {
		return fmt.Errorf("invalid key: %s", key)
	}
	c.ConfId = parts[0]
	c.Type = ProxieTesterType(parts[1])
	return nil
}

const ProxieKeyPrefix = "Proxie/"

type ProxieKey struct {
	ConfId     string
	SubName    string
	ProxieName string
}

func (p *ProxieKey) ToKey() []byte {
	return []byte(ProxieKeyPrefix + strings.Join([]string{p.ConfId, p.SubName, p.ProxieName}, "::"))
}

func (p *ProxieKey) FromKey(_key []byte) error {
	key := strings.TrimPrefix(string(_key), ProxieKeyPrefix)
	parts := strings.Split(key, "::")
	if len(parts) != 3 {
		return fmt.Errorf("invalid key: %s", key)
	}
	p.ConfId = parts[0]
	p.SubName = parts[1]
	p.ProxieName = parts[2]
	return nil
}

const ProxieResultKeyPrefix = "ProxieResult/"

type ProxieResultKey struct {
	ProxieKey
	Type ProxieTesterType
}

func (p *ProxieResultKey) ToKey() []byte {
	return []byte(ProxieResultKeyPrefix + strings.Join([]string{p.ConfId, p.SubName, p.ProxieName, string(p.Type)}, "::"))
}

func (p *ProxieResultKey) FromKey(_key []byte) error {
	key := strings.TrimPrefix(string(_key), ProxieResultKeyPrefix)
	parts := strings.Split(key, "::")
	if len(parts) != 4 {
		return fmt.Errorf("invalid key: %s", key)
	}
	p.ConfId = parts[0]
	p.SubName = parts[1]
	p.ProxieName = parts[2]
	p.Type = ProxieTesterType(parts[3])
	return nil
}
