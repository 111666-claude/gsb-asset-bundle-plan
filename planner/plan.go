// Package planner 按依赖闭包把资源包编成下载计划。
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

// 三色标记：0 未访问，1 在 DFS 栈上（灰），2 已处理完（黑）。
const (
	colorWhite = 0
	colorGray  = 1
	colorBlack = 2
)

type dfsFrame struct {
	name string
	dep  int
}

// Build 从 roots 出发编计划，每个桶容量是 budgetMB。
//
// 从 roots 按顺序做确定性的深度优先遍历，依赖按 Deps 中的声明顺序推进，
// 递归收集所有层级的依赖；指向灰色节点的边是一条回边，发现一次记一次
// Cycles；没登记过的名字记一次 Missing。收集到的包按发现顺序依次装桶，
// 桶内 SizeMB 之和不超过 budgetMB，单个包超预算时单独占一个桶。
// 整体代价是 O(包数 加 依赖边数)。
func Build(bundles []Bundle, roots []string, budgetMB int) Plan {
	plan := Plan{}

	// 建一次索引：每个包表条目只扫一遍，与 root 数无关。
	index := make(map[string]Bundle, len(bundles))
	for _, bundle := range bundles {
		plan.scanned++
		index[bundle.Name] = bundle
	}

	colors := map[string]int{}
	missing := map[string]bool{}
	order := make([]string, 0, len(bundles))
	stack := []dfsFrame{}

	recordMissing := func(name string) {
		if !missing[name] {
			missing[name] = true
			plan.Missing = append(plan.Missing, name)
		}
	}

	push := func(name string) {
		colors[name] = colorGray
		order = append(order, name)
		stack = append(stack, dfsFrame{name: name})
	}

	for _, root := range roots {
		bundle, ok := index[root]
		plan.scanned++
		if !ok {
			recordMissing(root)
			continue
		}
		if colors[root] != colorWhite {
			continue
		}
		push(root)

		for len(stack) > 0 {
			frame := &stack[len(stack)-1]
			if frame.dep >= len(bundle.Deps) {
				colors[frame.name] = colorBlack
				stack = stack[:len(stack)-1]
				if len(stack) > 0 {
					bundle = index[stack[len(stack)-1].name]
				}
				continue
			}
			dep := bundle.Deps[frame.dep]
			frame.dep++
			plan.scanned++

			depBundle, ok := index[dep]
			if !ok {
				recordMissing(dep)
				continue
			}
			switch colors[dep] {
			case colorGray:
				// 回边：发现一个环，环上的包已在遍历中，不重复入列。
				plan.Cycles++
			case colorBlack:
				// 已收集过，跳过即可。
			default:
				bundle = depBundle
				push(dep)
			}
		}
	}

	// 按发现顺序依次装桶；每个包最多出现一次，超预算的包单独占一桶。
	current := []string{}
	used := 0
	for _, name := range order {
		size := index[name].SizeMB
		if len(current) > 0 && used+size > budgetMB {
			plan.Buckets = append(plan.Buckets, current)
			current = []string{}
			used = 0
		}
		current = append(current, name)
		used += size
	}
	if len(current) > 0 {
		plan.Buckets = append(plan.Buckets, current)
	}

	return plan
}

// Scanned 是累计扫描的包表条目数（规模观测）。
func (p Plan) Scanned() int { return p.scanned }
