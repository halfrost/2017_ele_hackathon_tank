package metric

import (
	"math"
	"time"
)

const (
	// MaxSlot is the number for percentil index.
	MaxSlot = 100
	// BaseNumber is for pencentil index.
	BaseNumber = 0
)

var (
	powerOf4Index []int64
	buckets       []int64
)

func init() {
	powerOf4Index, buckets = initBuckets()
}

// GetTimespanKey  compute time span key.
func GetTimespanKey(t time.Time, aggratorTime time.Duration) int64 {
	if aggratorTime < time.Second {
		aggratorTime = time.Second
	}
	tt := t.UnixNano()
	agg := aggratorTime.Nanoseconds()
	return tt / agg
}

// GetPercentilIndex calc percentil index.
func GetPercentilIndex(val int64) int64 {
	amountIdx := PercentilIndex(val)
	if amountIdx > 0 && val == PercentilIndex(amountIdx-1) {
		amountIdx--
	}
	idx := amountIdx - BaseNumber
	if idx < 0 {
		return 0
	}
	if idx >= MaxSlot {
		return MaxSlot - 1
	}
	return idx
}

// PercentilIndex compute value index for histogram.
func PercentilIndex(val int64) int64 {
	if val <= 0 {
		return 0
	}
	if val <= 4 {
		return val
	}
	shift := BitLen(val) - 1
	prevPowerOf2 := (val >> shift) << shift
	prevPowerOf4 := prevPowerOf2
	denominator := 3
	if shift < 4 && shift%2 != 0 {
		shift--
		prevPowerOf4 = prevPowerOf2 >> 1
	} else if shift >= 4 && shift < 5 {
		denominator = 7
	} else if shift >= 5 && shift < 7 {
		denominator = 9
	} else if shift >= 7 && shift < 8 {
		denominator = 11
	} else if shift >= 8 && shift < 10 {
		denominator = 7
	} else if shift >= 10 {
		denominator = 5
	}
	base := prevPowerOf4
	delta := base / int64(denominator)
	offset := (val - base) / delta
	powerOf4Index := 0
	if shift < 4 {
		powerOf4Index = int(shift) / 2
	} else {
		powerOf4Index = int(shift) - 2
	}
	pos := offset + getPowerOf4Index(powerOf4Index)
	if pos >= int64(len(buckets)) {
		return int64(len(buckets) - 1)
	}
	if pos == getPowerOf4Index(powerOf4Index+1) {
		return pos
	}
	return pos + 1
}

func getPowerOf4Index(idx int) int64 {
	return powerOf4Index[idx]
}

// BitLen compute int64 most big 1 position.
func BitLen(x int64) (n uint) {
	for ; x >= 0x8000; x >>= 16 {
		n += 16
	}
	if x >= 0x80 {
		x >>= 8
		n += 8
	}
	if x >= 0x8 {
		x >>= 4
		n += 4
	}
	if x >= 0x2 {
		x >>= 2
		n += 2
	}
	if x >= 0x1 {
		n++
	}
	return
}

func initBuckets() ([]int64, []int64) {
	powerOf4Index := []int64{0}
	buckets := []int64{1, 2, 3}
	digits := 2
	denominator := 3
	exp := 2
	for exp < 64 {
		current := int64(1 << uint(exp))
		delta := current / int64(denominator)
		next := (current << uint(digits)) - delta
		powerOf4Index = append(powerOf4Index, int64(len(buckets)))
		for current <= next {
			buckets = append(buckets, current)
			current += delta
		}
		exp += digits
		if exp == 4 {
			digits = 1
			denominator = 7
		} else if exp == 5 {
			denominator = 9
		} else if exp == 7 {
			denominator = 11
		} else if exp == 8 {
			denominator = 7
		} else if exp == 10 {
			denominator = 5
		}
	}
	buckets = append(buckets, math.MaxInt64)
	return powerOf4Index, buckets
}
