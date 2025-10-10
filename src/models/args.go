package models

import (
	"encoding/json"
	"fmt"
)

type Args struct {
	Conf     Conf             `json:"conf"`
	Proxies  []map[string]any `json:"proxies"`
	Platform string           `json:"platform"`
	Context  struct {
		Source struct {
			Collection struct {
				Name                  string `json:"name"`
				DisplayName           string `json:"displayName"`
				Form                  string `json:"form"`
				Remark                string `json:"remark"`
				MergeSources          string `json:"mergeSources"`
				IgnoreFailedRemoteSub bool   `json:"ignoreFailedRemoteSub"`
				PassThroughUA         bool   `json:"passThroughUA"`
				Icon                  string `json:"icon"`
				IsIconColor           bool   `json:"isIconColor"`
				// Process               []struct {} `json:"process"`
				Subscriptions    []string `json:"subscriptions"`
				Tag              []any    `json:"tag"`
				SubscriptionTags []any    `json:"subscriptionTags"`
				DisplayName2     string   `json:"display-name"`
			} `json:"_collection"`
		} `json:"source"`
		Backend string   `json:"backend"`
		Version string   `json:"version"`
		Feature struct{} `json:"feature"`
		// Meta    struct {} `json:"meta"`
	} `json:"context"`
}

func (a *Args) GetProxieInfo(proxie map[string]any) *ProxieInfo {
	return &ProxieInfo{
		Id: ProxieKey{
			ConfId:     a.Conf.Id,
			SubName:    proxie["_subName"].(string),
			ProxieName: proxie["name"].(string),
		},
		Conf: &a.Conf,
	}
}

func (a *Args) UnmarshalJSON(data []byte) error {
	var obj map[string]json.RawMessage
	err := json.Unmarshal(data, &obj)
	if err != nil {
		return err
	}
	if v, ok := obj["conf"]; ok {
		if err := json.Unmarshal(v, &a.Conf); err != nil {
			return fmt.Errorf("conf: %w", err)
		}
	} else {
		a.Conf = *DefaultConf()
	}
	if v, ok := obj["proxies"]; ok {
		if err := json.Unmarshal(v, &a.Proxies); err != nil {
			return fmt.Errorf("proxies: %w", err)
		}
	}
	if v, ok := obj["args"]; ok {
		var rawArgs []json.RawMessage
		if err := json.Unmarshal(v, &rawArgs); err != nil {
			return fmt.Errorf("args: %w", err)
		}
		if len(rawArgs) >= 1 {
			_ = json.Unmarshal(rawArgs[0], &a.Proxies)
		}
		if len(rawArgs) >= 2 {
			_ = json.Unmarshal(rawArgs[1], &a.Platform)
		}
		if len(rawArgs) >= 3 {
			_ = json.Unmarshal(rawArgs[2], &a.Context)
		}
	}
	return nil
}
