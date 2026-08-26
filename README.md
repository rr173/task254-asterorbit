# task254-asterorbit · 近地小行星轨道观测残差归因服务

面向天体测量研究者，服务把观测弧段、测角记录、参考轨道与测站校准串成可复现的 O-C 残差归因闭环，并将最终轨道解发布为不可变快照。

## 标准验证

```bash
CGO_ENABLED=0 GOTOOLCHAIN=local go build ./...
go vet ./...
go test ./...
go run ./cmd/asterorbit --smoke-test
```

`--smoke-test` 会真实创建 SQLite 数据库，执行观测导入、轨道传播、残差计算、台站归因、校准收敛和快照发布，然后关闭并重开数据库验证恢复；退出码为 0 才表示通过。

## 服务启动

```bash
go run ./cmd/asterorbit --addr=:8080 --db=asterorbit.db
```

HTTP 接口统一使用 `/api` 前缀，覆盖弧段、台站、星表、观测、轨道、残差、归因、校准、快照和自检。
