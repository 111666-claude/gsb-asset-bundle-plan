// Package planner 按依赖闭包把资源包编成下载计划。
// 缺陷：闭包只取一层依赖、不检测依赖环、分桶完全不看预算、每次构建全扫包表。
package planner

// Bundle 是一个资源包。
type Bundle struct {
	Name   string
	SizeMB int
	Deps   []string
}

// Plan 是下载计划。
type Plan struct {
	Buckets  [][]string
	Missing  []string
	Cycles   int
	Deferred int
	scanned  int
}

// Build 从 roots 出发编计划，每个桶容量是 budgetMB。缺陷：只取一层依赖、不查环、不看预算、全扫包表。
func Build(bundles []Bundle, roots []string, budgetMB int) Plan {
	plan := Plan{}
	index := map[string]Bundle{}
	for _, bundle := range bundles {
		index[bundle.Name] = bundle
	}
	order := []string{}
	seen := map[string]bool{}
	for _, root := range roots {
		plan.scanned += len(bundles)
		if _, ok := index[root]; !ok {
			plan.Missing = append(plan.Missing, root)
			continue
		}
		if !seen[root] {
			seen[root] = true
			order = append(order, root)
		}
		for _, dep := range index[root].Deps {
			plan.scanned += len(bundles)
			if _, ok := index[dep]; !ok {
				plan.Missing = append(plan.Missing, dep)
				continue
			}
			if !seen[dep] {
				seen[dep] = true
				order = append(order, dep)
			}
		}
	}
	bucket := []string{}
	for _, name := range order {
		bucket = append(bucket, name)
	}
	if len(bucket) > 0 {
		plan.Buckets = append(plan.Buckets, bucket)
	}
	return plan
}

// Scanned 是累计扫描的包表条目数（规模观测）。
func (p Plan) Scanned() int { return p.scanned }
