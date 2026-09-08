package chalk_test

import (
	"testing"

	"github.com/chalk/chalk-go"
)

const benchmarkText = "the fox jumps over the lazy dog"

func benchmarkLevel(b *testing.B) chalk.Style {
	b.Helper()
	return chalk.MustNew(chalk.WithLevel(chalk.LevelTrueColor)).Root()
}

func BenchmarkOneStyle(b *testing.B) {
	root := benchmarkLevel(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = root.Red().Sprint(benchmarkText)
	}
}

func BenchmarkTwoStyles(b *testing.B) {
	root := benchmarkLevel(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = root.Blue().BgRed().Sprint(benchmarkText)
	}
}

func BenchmarkThreeStyles(b *testing.B) {
	root := benchmarkLevel(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = root.Blue().BgRed().Bold().Sprint(benchmarkText)
	}
}

func BenchmarkCachedOneStyle(b *testing.B) {
	red := benchmarkLevel(b).Red()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = red.Sprint(benchmarkText)
	}
}

func BenchmarkCachedTwoStyles(b *testing.B) {
	blueOnRed := benchmarkLevel(b).Blue().BgRed()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = blueOnRed.Sprint(benchmarkText)
	}
}

func BenchmarkCachedThreeStyles(b *testing.B) {
	blueOnRedBold := benchmarkLevel(b).Blue().BgRed().Bold()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = blueOnRedBold.Sprint(benchmarkText)
	}
}

func BenchmarkCachedOneStyleWithNewline(b *testing.B) {
	red := benchmarkLevel(b).Red()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = red.Sprint("the fox jumps\nover the lazy dog")
	}
}

func BenchmarkCachedOneStyleNestedIntersecting(b *testing.B) {
	root := benchmarkLevel(b)
	red := root.Red()
	nested := "the fox jumps" + root.Blue().Sprint("over the lazy dog") + "!"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = red.Sprint(nested)
	}
}

func BenchmarkCachedOneStyleNestedNonIntersecting(b *testing.B) {
	root := benchmarkLevel(b)
	bgRed := root.BgRed()
	nested := "the fox jumps" + root.Blue().Sprint("over the lazy dog") + "!"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = bgRed.Sprint(nested)
	}
}

func BenchmarkCachedOneStyleMultipleOperands(b *testing.B) {
	red := benchmarkLevel(b).Red()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = red.Sprint("the fox", "jumps over", "the lazy dog")
	}
}

func BenchmarkModelDownsampling(b *testing.B) {
	for _, level := range []struct {
		name  string
		level int
	}{
		{"basic", chalk.LevelBasic},
		{"ansi256", chalk.LevelAnsi256},
		{"truecolor", chalk.LevelTrueColor},
	} {
		root := chalk.MustNew(chalk.WithLevel(level.level)).Root()
		b.Run(level.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				sink = root.RGB(255, 136, 0).Sprint(benchmarkText)
			}
		})
	}
}

func BenchmarkDisabled(b *testing.B) {
	red := chalk.MustNew(chalk.WithLevel(chalk.LevelNone)).Root().Red()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = red.Sprint(benchmarkText)
	}
}

// sink keeps the compiler from eliminating the benchmarked calls.
var sink string
