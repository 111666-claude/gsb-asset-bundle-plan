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

const (
	stateUnvisited = 0
	stateOnStack   = 1
	stateDone      = 2
)

// frame 是迭代式 DFS 的栈帧：包名与下一条待处理的依赖下标。
type frame struct {
	name string
	next int
}

// Build 从 roots 出发编计划，每个桶容量是 budgetMB。
// 闭包递归收集所有层级的依赖；回边记一次 Cycles；
// 按 roots 顺序与依赖出现顺序装桶，桶容量不超过 budgetMB；
// 单个包超过预算时单独占一个桶（oversized，不计入 Deferred）。
func Build(bundles []Bundle, roots []string, budgetMB int) Plan {
	plan := Plan{}
	index := make(map[string]int, len(bundles))
	for position, bundle := range bundles {
		index[bundle.Name] = position
	}
	state := make([]int, len(bundles))
	order := []string{}
	missingSeen := map[string]bool{}
	addMissing := func(name string) {
		if !missingSeen[name] {
			missingSeen[name] = true
			plan.Missing = append(plan.Missing, name)
		}
	}
	for _, root := range roots {
		rootIndex, ok := index[root]
		if !ok {
			addMissing(root)
			continue
		}
		if state[rootIndex] != stateUnvisited {
			continue
		}
		state[rootIndex] = stateOnStack
		plan.scanned++
		order = append(order, root)
		stack := []frame{{name: root}}
		for len(stack) > 0 {
			top := &stack[len(stack)-1]
			bundle := bundles[index[top.name]]
			if top.next >= len(bundle.Deps) {
				state[index[top.name]] = stateDone
				stack = stack[:len(stack)-1]
				continue
			}
			dep := bundle.Deps[top.next]
			top.next++
			plan.scanned++
			depIndex, ok := index[dep]
			if !ok {
				addMissing(dep)
				continue
			}
			switch state[depIndex] {
			case stateUnvisited:
				state[depIndex] = stateOnStack
				plan.scanned++
				order = append(order, dep)
				stack = append(stack, frame{name: dep})
			case stateOnStack:
				plan.Cycles++
			}
		}
	}
	bucket := []string{}
	bucketSize := 0
	for _, name := range order {
		size := bundles[index[name]].SizeMB
		if len(bucket) > 0 && bucketSize+size > budgetMB {
			plan.Buckets = append(plan.Buckets, bucket)
			bucket = []string{}
			bucketSize = 0
		}
		bucket = append(bucket, name)
		bucketSize += size
		if bucketSize > budgetMB {
			plan.Buckets = append(plan.Buckets, bucket)
			bucket = []string{}
			bucketSize = 0
		}
	}
	if len(bucket) > 0 {
		plan.Buckets = append(plan.Buckets, bucket)
	}
	return plan
}

// Scanned 是累计扫描的包表条目数（规模观测）。
func (p Plan) Scanned() int { return p.scanned }
