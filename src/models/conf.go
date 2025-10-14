package models

import (
	"encoding/json"
	"log/slog"

	"github.com/ocyss/sub-store-lab/src/env"
	"github.com/rivo/uniseg"
)

type Conf struct {
	Id string `json:"id"`

	NoBeautifyNodes bool `json:"no_beautify_nodes"` // 是否禁用节点美化

	PurityCron string `json:"purity_cron"` // 纯净度测试 cron表达式
	SpeedCron  string `json:"speed_cron"`  // 速度/延迟测试 cron表达式

	SpeedTestUrl string `json:"speed_test_url"` // 测速测试URL

	MinSpeed        int `json:"min_speed"`        // 最低测速结果(KB/s)，低于此值舍弃，默认:256
	DownloadTimeout int `json:"download_timeout"` // 下载测试时间(秒)，与下载链接大小相关。默认:8
	DownloadMB      int `json:"download_mb"`      // 单节点测速下载数据大小(MB)限制，0为不限，默认:20

	PurityIconStr string `json:"purity_icon"`
	TypeIconStr   string `json:"type_icon"`

	PurityIcon []string `json:"-"`
	TypeIcon   []string `json:"-"`
}

func (c *Conf) Eq(other *Conf) bool {
	return c.Id == other.Id && c.SpeedTestUrl == other.SpeedTestUrl
}

var (
	PurityIconStr = "🖤🩵💙💛🧡❤️"
	TypeIconStr   = "🪨🏠🕋⚰️"
	PurityIcon    = splitEmoji(PurityIconStr)
	TypeIcon      = splitEmoji(TypeIconStr)
)

func DefaultConf() *Conf {
	return &Conf{
		PurityCron: "0 2 */3 * *", // 每3天的2点执行一次纯净度测试
		SpeedCron:  "0 3 * * *",   // 每天3点执行一次延迟测试

		// 默认测速URL
		SpeedTestUrl: "https://github.com/comfyanonymous/ComfyUI/releases/download/v0.3.57/ComfyUI_windows_portable_nvidia.7z",

		MinSpeed:        256,
		DownloadTimeout: 8,
		DownloadMB:      20,
		NoBeautifyNodes: env.Conf.DisableBeautify,
		PurityIconStr:   PurityIconStr,
		TypeIconStr:     TypeIconStr,
		PurityIcon:      PurityIcon,
		TypeIcon:        TypeIcon,
	}
}

func (c *Conf) UnmarshalJSON(data []byte) error {
	type Alias Conf
	defaults := DefaultConf()
	if err := json.Unmarshal(data, (*Alias)(defaults)); err != nil {
		return err
	}
	*c = Conf(*defaults)
	if v := splitEmoji(c.PurityIconStr); len(v) == len(PurityIcon) {
		c.PurityIcon = v
	} else {
		slog.Warn("purity icon length mismatch", "expected", len(PurityIcon), "got", len(v))
	}
	if v := splitEmoji(c.TypeIconStr); len(v) == len(TypeIcon) {
		c.TypeIcon = v
	} else {
		slog.Warn("type icon length mismatch", "expected", len(TypeIcon), "got", len(v))
	}
	return nil
}

func splitEmoji(s string) []string {
	var res []string
	gr := uniseg.NewGraphemes(s)
	for gr.Next() {
		res = append(res, gr.Str())
	}
	return res
}
