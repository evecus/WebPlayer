# 🎬 WebPlayer

一个运行在 Linux 上的单文件网页视频播放器服务器。

## 特性

- ✅ 单文件可执行，无任何外部依赖
- ✅ 支持 HLS（.m3u8）、MP4、WebM 等格式
- ✅ 导入 M3U / M3U8 播放列表文件
- ✅ 手动粘贴 M3U 内容 / 从 URL 拉取 M3U
- ✅ 播放列表持久化（重启保留）
- ✅ 播放历史记录（最近 50 条）
- ✅ 中英文切换
- ✅ 键盘快捷键
- ✅ 支持 linux/amd64 和 linux/arm64

---

## 构建

需要 Go 1.21+：

```bash
chmod +x build.sh
./build.sh
```

生成文件在 `dist/` 目录：
- `webplayer-linux-amd64`
- `webplayer-linux-arm64`

---

## 运行

```bash
chmod +x webplayer-linux-amd64

# 默认 8888 端口
./webplayer-linux-amd64

# 自定义端口
./webplayer-linux-amd64 -port 9090

# 自定义数据文件路径
./webplayer-linux-amd64 -port 8888 -data /var/data/vp.json
```

然后访问：`http://localhost:8888`（或你指定的端口）

---

## 使用

### 播放单个视频
在顶部地址栏粘贴视频地址，点击 Play 即可播放：
- `http://example.com/video.mp4`
- `http://example.com/live/stream.m3u8`

### 导入 M3U 播放列表
点击 **Import M3U** 按钮，支持三种方式：
1. 拖拽 / 选择 `.m3u` 文件
2. 粘贴 M3U 文本内容
3. 输入 M3U 文件的 HTTP 地址

### 手动添加频道
左侧列表点击 **Add** 按钮，填写频道名、地址等信息。

---

## 键盘快捷键

| 键 | 功能 |
|----|------|
| `Space` | 播放/暂停 |
| `←` | 后退 5 秒 |
| `→` | 前进 5 秒 |
| `↑` | 音量 +10% |
| `↓` | 音量 -10% |
| `F` | 全屏切换 |
| `M` | 静音切换 |

---

## 数据存储

所有播放列表和历史记录保存在 JSON 文件中（默认 `webplayer-data.json`，与可执行文件同目录）。

---

## API

服务同时提供 REST API，可供其他程序调用：

| Method | Path | 说明 |
|--------|------|------|
| GET | `/api/data` | 获取全部数据 |
| GET | `/api/playlists` | 获取所有播放列表 |
| POST | `/api/playlists` | 创建/更新播放列表 |
| DELETE | `/api/playlists/{id}` | 删除播放列表 |
| GET | `/api/history` | 获取历史记录 |
| POST | `/api/history` | 添加历史记录 |
| DELETE | `/api/history` | 清空历史记录 |
