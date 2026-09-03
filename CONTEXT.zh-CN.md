# RSSX

RSSX 是一个自托管的 RSS 阅读器：一个 Go 后端（`rssx-api`）负责把 RSS 源同步进存储并跟踪
已读状态，一个 Vue 前端（`rssx-ui`）负责展示。

> 本文件是 `CONTEXT.md` 的中文版本，内容与英文版保持一致。以英文版为准。

## 术语

**Feed（源）**：
RSSX 定期轮询的一个 RSS/Atom 源，例如 "InfoQ" 或 "InfoQ 中文"。用一个稳定的数字 id 标识。
Feed 的存在与谁订阅它无关。
_避免_：channel、source、site、频道、站点

**Subscription（订阅）**：
用户与某个 Feed 之间的关联。把 Feed 加入你的列表就创建一条 Subscription；移除它删除的是
Subscription，而不是 Feed。
_避免_：follow、watch、关注

**Purge（彻底删除）**：
把一个 Feed 整体删除——它的记录行、指向它的所有 Subscription，以及它在 Redis 里的全部
Article 和已读状态。与"移除 Subscription"不同，后者保留 Feed 本身。对应
`DELETE /feed/:id/purge` 和订阅源管理页的删除操作。
_避免_：只说 "delete / 删除" 而不说清删的是哪一个

**Article（文章）**：
从某个 Feed 抓取到的一条内容：标题、发布时间、原文 URL，以及 Feed 提供的正文（往往只有摘要）。
用每个 Feed 内部的 id 标识。
_避免_：news、item、entry、post、story、新闻、条目。（Go 的 `news` 包早于本术语表，
新命名应统一用 "Article"。）

**Sync（同步）**：
从某个 Feed 的原始 URL 拉取它当前的条目，并存下所有之前没见过的条目。按定时器运行，也可以
手动对单个 Feed 或全部 Feed 触发。
_避免_：refresh、poll、crawl、刷新、抓取

**Read boundary（已读边界）**：
按用户、按 Feed 记录：一个 Feed 文章历史中的某个位置，在这个位置之前的都视为已读。比边界更新的
文章除非被单独标记，否则都是未读。把边界往前推就是"全部标为已读"的实现方式。
_避免_：cursor、offset、watermark、游标、水位

**All view（All 视图）**：
左栏里的一个伪 Feed（id 为 `-1`），聚合所有 Subscription 的文章，而不代表某个真实的 Feed。
_避免_：inbox、everything、smart feed、收件箱
