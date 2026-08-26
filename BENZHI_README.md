基于 Go 实现的近地小行星轨道观测残差归因 Web 项目，一款后端服务，完成测角 O-C 残差计算、台站/星表/模型三向归因与不可变快照发布。

# BENZHI 评测说明

本项目为纯后端 Go 服务，对外暴露 `/api` 前缀的 HTTP 接口，使用 SQLite 持久化，
支持进程关闭后重新打开同一数据库恢复全部业务数据。

## 标准命令

```bash
CGO_ENABLED=0 GOTOOLCHAIN=local go build ./...
CGO_ENABLED=0 GOTOOLCHAIN=local go vet   ./...
CGO_ENABLED=0 GOTOOLCHAIN=local go test  ./...
go run ./cmd/asterorbit --smoke-test
go run ./cmd/asterorbit --addr :8080 --db asterorbit.db
```

- `--addr`：HTTP 监听地址，默认 `:8080`
- `--db`：SQLite 数据库文件路径，默认 `asterorbit.db`
- `--smoke-test`：不常驻；跑完端到端场景后关闭并重新打开数据库，退出码 0 表示通过

## 冒烟自测契约（--smoke-test）

创建临时数据库 → 导入观测弧段与测角 → 设置参考轨道并传播 → 计算 O-C 残差 →
按台站/星表/模型三向归因 → 校准台站零位使残差收敛 → 发布不可变残差快照 →
关闭并重新打开数据库，校验弧段/观测/轨道/残差/归因/快照全部持久化恢复后退出 0。

## Docker 构建与双架构验证

`Dockerfile` 与 `benzhi.Dockerfile` 内容完全一致。使用项目提供的
`build_benzhi_docker.sh` 构建评测镜像；Dockerfile 不声明端口，服务监听地址由
运行参数 `--addr` 指定。容器入口已经指向 `/app/asterorbit`，验证时不要再传二进制路径。

```bash
./build_benzhi_docker.sh task254-asterorbit:amd64 linux/amd64
docker run --rm task254-asterorbit:amd64 --smoke-test

./build_benzhi_docker.sh task254-asterorbit:arm64 linux/arm64
docker run --rm task254-asterorbit:arm64 --smoke-test

docker run --rm -P task254-asterorbit:amd64 --addr :8080 --db ./app.db
```

## 核心 API（`/api` 前缀）

- `POST /api/arcs` 创建观测弧段；`GET /api/arcs`、`GET /api/arcs/{id}`
- `POST /api/arcs/{id}/observations` 导入测角；`GET /api/arcs/{id}/observations`
- `POST /api/arcs/{id}/orbit` 设置参考轨道；`GET /api/arcs/{id}/orbit`
- `POST /api/arcs/{id}/compute` 计算残差；`GET /api/arcs/{id}/residuals`
- `POST /api/arcs/{id}/analyze` 三向归因；`GET /api/arcs/{id}/attribution`
- `POST /api/stations/{id}/calibrate` 校准台站零位
- `POST /api/arcs/{id}/snapshot` 发布不可变快照；`GET /api/arcs/{id}/snapshots`
- `GET /api/selfcheck`、`GET /api/health` 自省

## 业务不变量

- 弧段状态机：collecting→attributed→sealed（封存后不可再导入测角）。
- 轨道状态机：draft→active→superseded（同弧段仅一个 active）。
- 归因：pending→concluded，可按最新残差复算覆盖。
- 快照：published，内容哈希幂等，序号单调递增。
- 时间尺度统一到 TDB 后再做二体传播与测站地平几何。
