// Command asset-bundle-plan 跑资源包计划样例，并把计划写进文件。
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"example.com/bundles/planner"
)

func writePlan(path string, plan planner.Plan) error {
	lines := []string{}
	for index, bucket := range plan.Buckets {
		lines = append(lines, fmt.Sprintf("bucket-%d: %s", index, strings.Join(bucket, ",")))
	}
	lines = append(lines, fmt.Sprintf("missing=%d cycles=%d deferred=%d", len(plan.Missing), plan.Cycles, plan.Deferred))
	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o644)
}

// Run 执行一次命令行调用，返回退出码。
func Run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("asset-bundle-plan", flag.ContinueOnError)
	flags.SetOutput(stderr)
	sample := flags.String("sample", "closure", "closure / cycle / budget / work")
	out := flags.String("plan", "out/plan.txt", "计划文件路径")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	switch *sample {
	case "closure":
		plan := planner.Build([]planner.Bundle{
			{Name: "ui", SizeMB: 1, Deps: []string{"font"}},
			{Name: "font", SizeMB: 2, Deps: []string{"glyph"}},
			{Name: "glyph", SizeMB: 3},
		}, []string{"ui"}, 100)
		fmt.Fprintf(stdout, "bundles=%d\n", len(plan.Buckets[0]))
	case "cycle":
		plan := planner.Build([]planner.Bundle{
			{Name: "a", SizeMB: 1, Deps: []string{"b"}},
			{Name: "b", SizeMB: 1, Deps: []string{"a"}},
		}, []string{"a"}, 100)
		fmt.Fprintf(stdout, "cycles=%d\n", plan.Cycles)
	case "budget":
		plan := planner.Build([]planner.Bundle{
			{Name: "p1", SizeMB: 40}, {Name: "p2", SizeMB: 40}, {Name: "p3", SizeMB: 40},
		}, []string{"p1", "p2", "p3"}, 100)
		fmt.Fprintf(stdout, "buckets=%d\n", len(plan.Buckets))
	case "work":
		bundles := []planner.Bundle{}
		roots := []string{}
		for index := 0; index < 3000; index++ {
			name := fmt.Sprintf("b-%d", index)
			bundles = append(bundles, planner.Bundle{Name: name, SizeMB: 1})
			roots = append(roots, name)
		}
		plan := planner.Build(bundles, roots, 100)
		fmt.Fprintf(stdout, "scanned=%d\n", plan.Scanned())
	default:
		fmt.Fprintln(stderr, "需要 --sample closure|cycle|budget|work")
		return 2
	}
	plan := planner.Build([]planner.Bundle{{Name: "ui", SizeMB: 1}}, []string{"ui"}, 10)
	if err := writePlan(*out, plan); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

func main() {
	os.Exit(Run(os.Args[1:], os.Stdout, os.Stderr))
}
