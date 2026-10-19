# 百万邮箱部署记录 2026-10-06

设备：ADMIN-015 / app-win-01。

## 当前入口

- 工作站：http://127.0.0.1:3040
- 109：https://pxy.xyz.hr/apps/mail/
- 204：http://43.167.9.204/apps/mail/

源码：D:/x1/x2/PXYMAIL。按用户最新要求，运行程序、数据库、配置和日志改为 E:/pxy-runtime/Mail。D 盘此前的运行数据保留为停用备份，迁移过程和结果见 E:/pxy-runtime/Mail/runtime-migration-result.json。

## 完成结果

- 保留 19 个 SMTP 配置、17 个模板、11,311 条联系人记录和 46 个历史任务。
- Mail 历史记录保留 SMTP 已接收 91 条、发送失败 1 条；另有 10 条未发送记录属于已关闭任务，不恢复自动发送。
- 发件账号、旧发送/退信记录、退订配置、原始 CSV 和模板文件已独立保存到 configuration；发送脚本使用此目录，不依赖 YXHK 本地目录。
- 开机及登录启动任务：PXY-MailNative-Start，运行身份 SYSTEM。
- YXHK 四个服务已停止，启动类型 Manual。源码新增待删除说明.md，原数据保留备用；2026-10-20 用户已确认将 YXHK 源码移入 `D:/x1/x2/PXYLH/LHBAT/待删除/YXHK-20261020` 归档，营销统一以 Mail 为准。
- 尚未确定名称的第三个邮件项目未删除。

## 本次验证

- 前端 rsbuild 构建成功，Go Windows 构建成功。
- PowerShell 脚本语法、Caddy 配置、109/204 Nginx 配置校验通过。
- 本机及两个公网入口健康检查返回 BillionMail / native / ok。
- 三个入口均使用浏览器登录并刷新概览；原生统计和 4 张图表正常，未发现 JavaScript 异常或失败 HTTP 请求。
- 模板和发送任务接口正常，默认今天无新发送时显示 0；查看旧发送记录需选择对应日期。
- 本次未创建发送任务，未发送邮件。

## 限制

工作站必须开机，109/204 只代理请求，不承载邮件数据库。SMTP 已接收不代表最终送达或进入收件箱。

## 收件箱和客户回复接入

- 用户确认需要直接查看原有 126/163 收件箱和客户回复；Windows 版已加入“收件箱 / 客户回复”菜单。
- 接入现有 16 个网易邮箱（8 个 126、8 个 163），复用数据库中的加密 SMTP 授权码，不新增明文密码文件。
- 使用 IMAP 993 TLS 和网易客户端 ID，每次以只读方式打开 INBOX；后台按账号轮询，每账号同步间隔至少 15 分钟，账号间登录至少 20 秒。
- 同步最近 100 封列表，批量缓存最近 10 封正文；其他正文按需读取。处理标记仅保存在本项目的 E 盘数据库。
- 邮件正文支持纯文本和 HTML 预览，附件列出名称；HTML 禁用脚本、远程图片及跳转，超过 1 MB 的邮件和附件可打开原邮箱查看。
- 分类依据迁移的发送历史及当前项目发送记录；客户回复和退信是辅助分类，不自动更改拒发名单或最终送达统计。
- 首次同步已获取 105 封列表，包含 2 封客户回复及 11 封退信；zmx02w@126.com 初次读取部分正文失败，增加批量读取及正文缓存独立状态处理。正文重试成功，最终 16 个账号均连接正常、无连接错误，105 封正文全部缓存。
- 三个入口登录、刷新、列表、正文预览均正常；820 像素平板宽度无页面横向溢出，未发现 JavaScript 异常或失败 HTTP 请求。
- 本次未创建发送任务；仍为 0 个未完成且未暂停的发送任务，SMTP 已接收历史仍为 91 条。

收件运行明细保存在 E:/pxy-runtime/Mail/inbox-status.json 和 inbox-browser-status.json。操作说明见 deploy/windows/README.md。

## 源码与部署

PXYOPS base_commit：4cea586。最终提交：e316883，已推送 origin/main 并部署工作站 Caddy 以及 109/204 Nginx。业务修改仅限 Mail；YXHK 仅新增待删除标记。Mail 当前不是 Git 仓库，修改保存在本机源码目录。部署入口和路径见 deploy/windows/README.md。

验证明细保存于 E:/pxy-runtime/Mail/deployment-browser-check.json、workstation-setup-result.json、yxhk-standby-result.json、autostart-validation.json。

## E 盘运行目录修正

2026-10-06 19:33 按用户最终要求完成 D -> E 冷迁移。当前程序为 E:/pxy-runtime/Mail/core/billionmail.exe，实际 PostgreSQL data_directory 为 E:/pxy-runtime/Mail/postgresql；启动和发送脚本、退订配置路径及开机任务均指向 E 盘。

迁移前后核对一致：19 个 SMTP 配置、17 个模板、11,311 个联系人、46 个发送任务、91 条 SMTP 已接收记录、16 个收件账号、105 封邮件列表及正文。

旧 E 盘副本保存在 E:/pxy-runtime/Mail-before-rehome-20261006-193215。D 盘此前运行数据已停用并改名为 D:/pxy-runtime/Mail-inactive-after-E-move-20261006-193215，D:/pxy-runtime/Mail 不再作为运行目录。源码及 MAILBAT 继续保留在 D:/x1/x2/PXYMAIL。

本机、109 和 204 健康检查均为 BillionMail / native / ok；迁移结果保存在 E:/pxy-runtime/Mail/runtime-migration-result.json。
