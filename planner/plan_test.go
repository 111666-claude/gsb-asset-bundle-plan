package planner

import "testing"

func TestBuildSingleBundle(t *testing.T) {
	plan := Build([]Bundle{{Name: "a", SizeMB: 1}}, []string{"a"}, 10)
	if len(plan.Buckets) != 1 || len(plan.Buckets[0]) != 1 {
		t.Fatalf("单个包应该编进一个桶，得到 %+v", plan.Buckets)
	}
}

func TestBuildReportsMissing(t *testing.T) {
	plan := Build([]Bundle{{Name: "a", SizeMB: 1, Deps: []string{"ghost"}}}, []string{"a"}, 10)
	if len(plan.Missing) != 1 || plan.Missing[0] != "ghost" {
		t.Fatalf("未知依赖应该记进 Missing，得到 %+v", plan.Missing)
	}
}

func TestBuildKeepsInputOrder(t *testing.T) {
	plan := Build([]Bundle{{Name: "a", SizeMB: 1}, {Name: "b", SizeMB: 1}}, []string{"b", "a"}, 10)
	if plan.Buckets[0][0] != "b" {
		t.Fatalf("桶内顺序应该跟 roots 顺序一致，得到 %+v", plan.Buckets[0])
	}
}
