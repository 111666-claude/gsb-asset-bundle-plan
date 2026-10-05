# asset-bundle-plan

资源包计划：按依赖闭包收集资源包，再按容量预算分桶，并写出一份计划文件，只用 Go 标准库。

```
go run ./cmd/asset-bundle-plan --sample=closure
go run ./cmd/asset-bundle-plan --sample=cycle
go run ./cmd/asset-bundle-plan --sample=budget
go run ./cmd/asset-bundle-plan --sample=work
go run ./cmd/asset-bundle-plan --sample=closure -plan out/plan.txt
go test ./...
go vet ./...
```

## 口径（README 为准）

- **依赖闭包**：从 `roots` 出发递归收集所有层级的依赖；没登记过的依赖记进 `Missing`。
- **依赖环**：闭包推进过程中发现环就记一次 `Cycles`，环里的包不重复进桶。
- **分桶**：按 `roots` 顺序与依赖出现顺序依次装桶，每个桶的 `SizeMB` 之和不超过 `budgetMB`；
  单个包超过预算时它单独占一个桶并计入 `Deferred` 之外的 oversized 提示。
- **确定性**：同样输入重复构建结果逐字节一致，不依赖 map 迭代顺序。
- **不变量**：每个包在所有桶里最多出现一次；桶内顺序与出现顺序一致。
- **代价**：构建是 O(包数 加 依赖边数)，`Scanned` 不随包数乘 root 数增长；10 万包每秒 10 万次构建。
- **计划文件**：`-plan` 指定的文件每行一个桶 `bucket-N: a,b,c`，最后一行汇总 `missing=.. cycles=.. deferred=..`。

## 输出契约（不改格式）

```
bundles=..
cycles=..
buckets=..
scanned=..
bucket-0: ..        （计划文件内容）
missing=.. cycles=.. deferred=..
```

## 目录

```
planner/plan.go                  闭包、环检测与分桶
cmd/asset-bundle-plan/main.go    命令行入口与计划文件
planner/plan_test.go             构建用例
```
