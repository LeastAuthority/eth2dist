package eth2dist

import (
	"bytes"
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"testing"

	"github.com/montanaflynn/stats"
)

func UniformValidatorsState(total, n int) State {
	assert(total%n == 0)

	state := State{make([]Validator, n)}
	for i := range state.Validators {
		state.Validators[i].EffectiveBalance = total / n
	}
	return state
}

func seq(from, to int) []ValidatorIndex {
	out := make([]ValidatorIndex, to-from)
	for i := from; i < to-from; i++ {
		out[i-from] = ValidatorIndex(i)
	}
	return out
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
			for i := range counts {
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

		rmean, err := stats.Mean(rfcnts)
		t.Log("rmean:", rmean, err)
		rstddev, err := stats.StandardDeviation(rfcnts)
		t.Log("rstddev:", rstddev, err)
	}

}

func TestComputeProposerIndex(t *testing.T) {
	for _, setSize := range []int{32} {
		for _, samples := range []int{16 * 1024} {
			t.Run(fmt.Sprintf("sz:%d/smpls:%d", setSize, samples), getTest(setSize, samples))
		}
	}
}
