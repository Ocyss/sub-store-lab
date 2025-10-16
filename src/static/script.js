async function operator(...args) {
    // 根据lab后端进行调整: host.docker.internal:8000, sub-store-lab:8000
    const resp = await fetch("http://127.0.0.1:8000", {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
        },
        body: JSON.stringify({
            conf: {
                // conf为可选项
                // id: "", // 指定当前订阅id
                // purity_cron: "0 2 */3 * *",// 纯净度测试 cron表达式
                // speed_cron: "0 3 * * *",// 速度/延迟测试 cron表达式
                // speed_test_url: "", // 测速下载Url
                // min_speed: "256",// 最低测速结果(KB/s)，低于此值舍弃，默认:256
                // download_timeout: "8",// 下载测试时间(秒)，与下载链接大小相关。默认:8
                // download_mb: "20",// 单节点测速下载数据大小(MB)限制，0为不限，默认:20
                // keyword_keep: "", // 关键词保留，| 竖线分割, 示例: 福利|家宽|流媒
                // purity_icon:"🖤|🩵|💙|💛|🧡|❤️", // 数量要严格一致并用竖线|分割，避免emoji分割错误
                // type_icon:"🪨|🏠|🕋",
            },
            args
        }),
    }).then(r => r.json())
    return resp
}