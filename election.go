package eth2dist // import "github.com/keks/eth2dist"

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

/*
def compute_shuffled_index(index: ValidatorIndex, index_count: uint64, seed: Bytes32) -> ValidatorIndex:
    assert index < index_count

    # Swap or not (https://link.springer.com/content/pdf/10.1007%2F978-3-642-32009-5_1.pdf)
    # See the 'generalized domain' algorithm on page 3
    for current_round in range(SHUFFLE_ROUND_COUNT):
        pivot = bytes_to_int(hash(seed + int_to_bytes(current_round, length=1))[0:8]) % index_count
        flip = ValidatorIndex((pivot + index_count - index) % index_count)
        position = max(index, flip)
        source = hash(seed + int_to_bytes(current_round, length=1) + int_to_bytes(position // 256, length=4))
        byte = source[(position % 256) // 8]
        bit = (byte >> (position % 8)) % 2
        index = flip if bit else index

    return ValidatorIndex(index)

def compute_proposer_index(state: BeaconState, indices: Sequence[ValidatorIndex], seed: Bytes32) -> ValidatorIndex:
    assert len(indices) > 0
    MAX_RANDOM_BYTE = 2**8 - 1
    i = 0
    while True:
        candidate_index = indices[compute_shuffled_index(ValidatorIndex(i % len(indices)), len(indices), seed)]
        random_byte = hash(seed + int_to_bytes(i // 32, length=8))[i % 32]
        effective_balance = state.validators[candidate_index].effective_balance
        if effective_balance * MAX_RANDOM_BYTE >= MAX_EFFECTIVE_BALANCE * random_byte:
            return ValidatorIndex(candidate_index)
        i += 1
*/

type State struct {
	Validators []Validator
}

type Validator struct {
	EffectiveBalance int
}

type ValidatorIndex uint64

const (
	ShuffleRoundCount   = 90
	MaxEffectiveBalance = 32 * 1000 * 1000 * 1000
)

func ComputeShuffledIndex(idx ValidatorIndex, idxCnt uint64, seed [32]byte) ValidatorIndex {
	assert(uint64(idx) < idxCnt)

	for round := 0; round < ShuffleRoundCount; round++ {
		pivot := uint64hash(seed, round)
		flip := ValidatorIndex((pivot + idxCnt - uint64(idx)) % idxCnt)
		pos := max(idx, flip)
		src := hash(seed, round, pos/256)
		bite := src[(pos%256)/8]
		byt := (bite >> (pos % 8)) % 2
		if byt == 1 {
			idx = flip
		}
	}
	return idx
}

func ComputeProposerIndex(state State, indices []ValidatorIndex, seed [32]byte) ValidatorIndex {
	assert(len(indices) > 0)

	const max = 255
	var i = 0

	for {
		var (
			shufIdx = ComputeShuffledIndex(ValidatorIndex(i%len(indices)), uint64(len(indices)), seed)
			candIdx = indices[shufIdx]
			r       = hash(seed, i/32)[i%32]
			eff     = state.Validators[candIdx].EffectiveBalance
		)
		if eff*max >= MaxEffectiveBalance*int(r) {
			return candIdx
		}
		i += 1
	}
}

func assertf(cond bool, format string, data ...interface{}) {
	if !cond {
		panic(fmt.Sprintf("assertion failed - "+format, data...))
	}
}

func assert(cond bool) {
	if !cond {
		panic("assertion failed")
	}
}

func max(l, r ValidatorIndex) ValidatorIndex {
	if l > r {
		return l
	}

	return r
}

func uint64hash(vs ...interface{}) uint64 {
	var (
		out uint64
		h   = hash(vs...)
	)

	for i := 0; i < 8; i++ {
		out += uint64(h[i])
		out = out << 8
	}

	return out
}

func hash(vs ...interface{}) []byte {
	h := sha256.New()
	e := json.NewEncoder(h)

	for _, v := range vs {
		err := e.Encode(v)
		assertf(err == nil, "encode error: %q", err)
	}

	return h.Sum(nil)
}
