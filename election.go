package eth2dist // import "github.com/LeastAuthority/eth2dist"

import (
	"fmt"

	"crypto/sha256"
	//"github.com/codahale/blake2"
	//"lukechampine.com/blake3"
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

	// buf is used as a hashing buffer.
	// initially tests took ages so I optimized this code a bit.
	// it's not perfectly avoiding collisions, but I think that
	// is okay for testing uniformity.
	// buf layout: b32:seed | b1:round# | b1: sel
	// sel is needed to get different hashes. in the real protocol,
	// we hash in some other value but that is not too important here.
	// writing to buffers directly instead of doing lazy JSON encoding
	// or fprint-ing into the hash resulted in 10x performance boost.
	var buf = make([]byte, len(seed)+2)
	copy(buf, seed[:])

	for round := 0; round < ShuffleRoundCount; round++ {
		buf[32] = byte(round)
		buf[33] = 0
		pivot := uint64hash(buf)
		flip := ValidatorIndex((pivot + idxCnt - uint64(idx)) % idxCnt)
		pos := max(idx, flip)
		buf[33] = 1
		src := hash(buf)
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
	var (
		i      = 0
		hashes = make([][]byte, 0, 500) // allocate a lot of memory up front
		buf    = make([]byte, len(seed)+1)
	)

	copy(buf, seed[:])

	for {
		// only compute the hash if we haven't done so yet
		if i/32 > len(hashes)-1 {
			buf[32] = byte(i / 32)
			hashes = append(hashes, hash(buf))
		}

		var (
			shufIdx = ComputeShuffledIndex(ValidatorIndex(i%len(indices)), uint64(len(indices)), seed)
			candIdx = indices[shufIdx]
			r       = hashes[i/32][i%32]
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

func uint64hash(buf []byte) uint64 {
	var (
		out uint64
		h   = hash(buf)
	)

	for i := 0; i < 8; i++ {
		out += uint64(h[i])
		out = out << 8
	}

	return out
}

//var blakeCfg = blake2.Config{Size: 32}

func hash(buf []byte) []byte {
	//h := blake2.NewBlake2B()
	//h := blake2.New(&blakeCfg)
	h := sha256.New()
	//h := blake3.New(32, nil)
	//e := json.NewEncoder(h)

	h.Write(buf)
	return h.Sum(nil)
}
