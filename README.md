# Sub-Store-Lab 🧪

> ⚠️ **注意**：本项目正在开发中，尚未经过长久测试，不建议使用。
>
> ℹ️ **说明**：IP纯净度的检测涉及多个api提供商，需要自备key，并且因为没有付费检测效果不理想，但作为参考还是可以的。
>
> 🛑 **免责声明**：本项目仅供学习和研究使用，开发者不对因使用本项目而产生的任何直接或间接损失负责。请遵守当地法律法规，勿用于任何非法用途。

## 📚 项目介绍

Sub-Store-Lab 是一个用于订阅节点管理、测试和美化的工具框架。它通过与 Sub-Store 解耦的方式，提供节点美化、排序以及性能测试功能。

推荐客户端: [clash-party](https://github.com/mihomo-party-org/clash-party) 自带sub-store前端

拉取平台设为JSON即可强制后台刷新

### ✨ 主要功能

- 🚀 **速率测试** - 测试节点的下载速度和延迟
- 🔍 **纯净度测试** - 检测节点的质量和可用性
- 🎨 **节点美化** - 优化节点名称显示, 保留倍率信息
- 📊 **智能排序** - 根据测试结果对节点进行排序
  - 国家/地区平均延迟进行升序
  - 订阅通过显示名称 `name:num` 语法进行升序
- ⚙️ **高度可配置** - 通过 conf 选项自定义各种参数

## 📸 效果展示

> 先这样吧，typeicon有点难看。。不过配置中可以自定义, 有合适的icon可以提issue

![img](docs/image/QQ20251010-190427.png)

## 📚 说明

### 🛠️部署流程

<!-- todo 优化部署教程 -->
建议使用 docker-compose 进行部署，参考如下配置：

- [docker-compose示例](docker-compose.yml)
- [docker-compose-caddy-proxy示例](docker-compose-caddy.yml)

后端地址则为 service_name:8000, `http://sub-store-lab:8000`

### 🛠️使用方法

通过脚本操作实现与 Sub-Store 框架的解耦，示例代码：

完整配置项可以参考 [script.js](src/script.js)，尽量保证和conf的同步更新

```javascript
async function operator(...args) {
    const resp = await fetch("http://127.0.0.1:8000", {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
        },
        body: JSON.stringify({
            conf: {
                // 可选配置项
            },
            args
        }),
    }).then(r => r.json())
    return resp
}
```

## 📝 鸣谢

不分先后

- [VPS IP 质量检测完全指南：从小白到精通的实用教程 - idcflare.com](https://idcflare.com/t/topic/18792)
- [IP 质量 - 快速排查清单 - linux.do](https://linux.do/t/topic/997322)

- [metacubex/mihomo](https://github.com/metacubex/mihomo)
- [sub-store-org/Sub-Store](https://github.com/sub-store-org/Sub-Store)
- [beck-8/subs-check](https://github.com/beck-8/subs-check)
- [bestruirui/BestSub](https://github.com/bestruirui/BestSub)
- [oneclickvirt/ecs](https://github.com/oneclickvirt/ecs)
- [xykt/IPQuality](https://github.com/xykt/IPQuality)

- [AbuseIPDB](https://www.abuseipdb.com/)
- [IPAPI](https://ipapi.co/)
- [IPData](https://ipdata.co/)
- [IPinfo](https://ipinfo.io/)
- [IPQualityScore](https://www.ipqualityscore.com/)
- [IPRegistry](https://ipregistry.co/)

- ipify.org, amazonaws.com, ifconfig.me, ident.me, icanhazip.com, api.ip.sb, ipinfo.io, ipapi.co

## ⭐ Stars 历史

[![Stargazers over time](https://starchart.cc/ocyss/sub-store-lab.svg?variant=adaptive)](https://starchart.cc/ocyss/sub-store-lab)
