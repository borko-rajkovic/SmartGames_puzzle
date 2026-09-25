package utility

type Stats struct {
	Total   uint64
	Visited uint64
	Skipped uint64
}

func factorial(n uint64) uint64 {
	if n == 0 || n == 1 {
		return 1
	}

	result := uint64(1)

	for i := uint64(2); i <= n; i++ {
		result *= i
	}

	return result
}

type VisitResult struct {
	ShouldSkipPermutations bool
	ShouldAbort            bool
}

func PermuteDigits(digits []rune, visit func(prefix string, progress float64) VisitResult) Stats {
	used := make([]bool, len(digits))
	result := make([]rune, len(digits))

	stats := Stats{
		Total: factorial(uint64(len(digits))),
	}

	var generate func(pos int, abort bool) bool

	generate = func(pos int, abort bool) bool {
		if abort {
			return true
		}

		prefix := string(result[:pos])

		progress := (float64(stats.Visited) + float64(stats.Skipped)) / float64(stats.Total) * 100
		progress = float64(int(progress*100)) / 100

		if pos != len(result) {
			visitResult := visit(prefix, progress)

			if visitResult.ShouldAbort {
				return true
			}

			if visitResult.ShouldSkipPermutations {
				stats.Skipped += factorial(uint64(len(result) - pos))
				return false
			}
		}

		if pos == len(result) {
			stats.Visited++
			progress := (float64(stats.Visited) + float64(stats.Skipped)) / float64(stats.Total) * 100
			progress = float64(int(progress*100)) / 100
			visitResult := visit(string(result), progress)
			return visitResult.ShouldAbort
		}

		for i, digit := range digits {
			if used[i] {
				continue
			}

			used[i] = true
			result[pos] = digit

			result := generate(pos+1, false)
			if result {
				return true
			}

			used[i] = false
		}

		return false
	}

	generate(0, false)

	return stats
}
