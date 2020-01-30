package eth2dist

import (
	"bytes"
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"testing"
	"math"

	"github.com/montanaflynn/stats"
)

// parameters
var (
	setSizes     = []int{8, 32}
	sampleCounts = []int{1024, 32 * 1024}
)

// UniformValudatorsState returns a state with total effective balance
// `total` and n validators. The total balance is evenly allocated to the
// validators.
func UniformValidatorsState(total, n int) State {
	assert(total%n == 0)

	state := State{make([]Validator, n)}
	for i := range state.Validators {
		state.Validators[i].EffectiveBalance = total / n
	}
	return state
}

// Null hypothesis, there is no statistically significant difference between
// observed sample distribution and expected random selection.
func ChaiSquaredTest(a []float64, samples int) float64 {
	expected := float64(samples / len(a))
	sum := 0.0
	for i := 0; i < len(a); i++{
		sum += math.Pow(a[i] - expected, 2) / float64(expected)
	}
	return sum
}

func getTest(setSize, samples int) func(*testing.T) {
	var (
		state = UniformValidatorsState(1024, 64)
		seed  [32]byte
	)

	return func(t *testing.T) {
		var (
			cores   = runtime.NumCPU()
			wg      sync.WaitGroup
			chunksz = samples / cores

			// per thread
			counts  = make([][]int, cores)
			rcounts = make([][]int, cores)
		)

		fmt.Println("cores", cores)
		wg.Add(cores)

		for j := 0; j < cores; j++ {
			go func(j int) {
				var (
					buf bytes.Buffer
				)

				counts[j] = make([]int, setSize)
				rcounts[j] = make([]int, setSize)

				start := j * chunksz
				end := (j + 1) * chunksz
				for i := start; i < end; i++ {
					// prepare seed
					fmt.Fprint(&buf, i)
					copy(seed[:], buf.Bytes())
					buf.Reset()

					// add result of "real" computation
					idx := ComputeProposerIndex(state, seq(0, setSize), seed)
					counts[j][int(idx)]++

					// add random value
					ridx := rand.Intn(setSize)
					rcounts[j][ridx]++

					// print progress
					if i%32 == 0 {
						l := len(fmt.Sprint(samples))
						format := fmt.Sprintf("%%0%dd/%%0%dd %%x\n", l, l)
						fmt.Printf(format, i, samples, seed)
					}
				}
				wg.Done()
			}(j)
		}

		wg.Wait()

		var (
			fcnts  = make(stats.Float64Data, len(counts[0]))
			rfcnts = make(stats.Float64Data, len(counts[0]))
		)

		for j := 0; j < cores; j++ {
			for i := range counts[j] {
				fcnts[i] += float64(counts[j][i])
				rfcnts[i] += float64(rcounts[j][i])
			}
		}

		t.Log("real:", fcnts)
		t.Log("rand:", rfcnts)

		mean, err := stats.Mean(fcnts)
		t.Log("mean:", mean, err)
		stddev, err := stats.StandardDeviation(fcnts)
		t.Log("stddev:", stddev, err)
		chai := ChaiSquaredTest(fcnts, samples)
		t.Log("chai real:", chai)
		if len(fcnts) == 8 {
			if chai < 14.07 {
				t.Log("95 percent confident in null hypothesis")
			} else {
				t.Log("There is a statistical difference between observed and expected")
			}
		} else if len(fcnts) == 32 {
			if chai < 43.77 {
				t.Log("95 percent confident in null hypothesis")
			} else {
				t.Log("There is a statistical difference between observed and expected")
			}			
		}

		rmean, err := stats.Mean(rfcnts)
		t.Log("rmean:", rmean, err)
		rstddev, err := stats.StandardDeviation(rfcnts)
		t.Log("rstddev:", rstddev, err)
		rchai := ChaiSquaredTest(rfcnts, samples)
		t.Log("chai rand:", rchai)
		if len(rfcnts) == 8 {
			if rchai < 14.07 {
				t.Log("95 percent confident in null hypothesis")
			} else {
				t.Log("There is a statistical difference between observed and expected")
			}
		} else if len(rfcnts) == 32 {
			if rchai < 43.77 {
				t.Log("95 percent confident in null hypothesis")
			} else {
				t.Log("There is a statistical difference between observed and expected")
			}			
		}
	}

}

func TestComputeProposerIndex(t *testing.T) {
	for _, setSize := range setSizes {
		for _, samples := range sampleCounts {
			t.Run(fmt.Sprintf("sz:%d/smpls:%d", setSize, samples), getTest(setSize, samples))
		}
	}
}

func seq(from, to int) []ValidatorIndex {
	out := make([]ValidatorIndex, to-from)
	for i := from; i < to-from; i++ {
		out[i-from] = ValidatorIndex(i)
	}
	return out
}
