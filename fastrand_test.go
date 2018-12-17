package fastrand

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"unsafe"

	valyalarand "github.com/valyala/fastrand"
)

func TestRace(t *testing.T) {
	var BenchSink uint32
	var wg sync.WaitGroup
	wg.Add(500)
	for i := 0; i < 500; i++ {
		go func() {
			defer wg.Done()
			j := Uint32()
			atomic.AddUint32(&BenchSink, j)
		}()
	}
}

// BenchSink prevents the compiler from optimizing away benchmark loops.
var BenchSink uint32

func BenchmarkFastRandUint32n(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		s := uint32(0)
		for pb.Next() {
			s += Uint32n(1e6)
		}
		atomic.AddUint32(&BenchSink, s)
	})
}

func BenchmarkUint32n(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		s := uint32(0)
		for pb.Next() {
			s += valyalarand.Uint32n(1e6)
		}
		atomic.AddUint32(&BenchSink, s)
	})
}

func BenchmarkRNGUint32n(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		var r valyalarand.RNG
		s := uint32(0)
		for pb.Next() {
			s += r.Uint32n(1e6)
		}
		atomic.AddUint32(&BenchSink, s)
	})
}

func BenchmarkRNGUint32nWithLock(b *testing.B) {
	var r valyalarand.RNG
	var rMu sync.Mutex
	b.RunParallel(func(pb *testing.PB) {
		s := uint32(0)
		for pb.Next() {
			rMu.Lock()
			s += r.Uint32n(1e6)
			rMu.Unlock()
		}
		atomic.AddUint32(&BenchSink, s)
	})
}

func BenchmarkRNGUint32nArray(b *testing.B) {
	var rr [64]struct {
		r  valyalarand.RNG
		mu sync.Mutex

		// pad prevents from false sharing
		pad [64 - (unsafe.Sizeof(valyalarand.RNG{})+unsafe.Sizeof(sync.Mutex{}))%64]byte
	}
	var n uint32
	b.RunParallel(func(pb *testing.PB) {
		s := uint32(0)
		for pb.Next() {
			idx := atomic.AddUint32(&n, 1)
			r := &rr[idx%uint32(len(rr))]
			r.mu.Lock()
			s += r.r.Uint32n(1e6)
			r.mu.Unlock()
		}
		atomic.AddUint32(&BenchSink, s)
	})
}

func BenchmarkMathRandInt31n(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		s := uint32(0)
		for pb.Next() {
			s += uint32(rand.Int31n(1e6))
		}
		atomic.AddUint32(&BenchSink, s)
	})
}

func BenchmarkMathRandRNGInt31n(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		r := rand.New(rand.NewSource(42))
		s := uint32(0)
		for pb.Next() {
			s += uint32(r.Int31n(1e6))
		}
		atomic.AddUint32(&BenchSink, s)
	})
}

func BenchmarkMathRandRNGInt31nWithLock(b *testing.B) {
	r := rand.New(rand.NewSource(42))
	var rMu sync.Mutex
	b.RunParallel(func(pb *testing.PB) {
		s := uint32(0)
		for pb.Next() {
			rMu.Lock()
			s += uint32(r.Int31n(1e6))
			rMu.Unlock()
		}
		atomic.AddUint32(&BenchSink, s)
	})
}

func BenchmarkMathRandRNGInt31nArray(b *testing.B) {
	var rr [64]struct {
		r  *rand.Rand
		mu sync.Mutex

		// pad prevents from false sharing
		pad [64 - (unsafe.Sizeof(valyalarand.RNG{})+unsafe.Sizeof(sync.Mutex{}))%64]byte
	}
	for i := range rr {
		rr[i].r = rand.New(rand.NewSource(int64(i)))
	}

	var n uint32
	b.RunParallel(func(pb *testing.PB) {
		s := uint32(0)
		for pb.Next() {
			idx := atomic.AddUint32(&n, 1)
			r := &rr[idx%uint32(len(rr))]
			r.mu.Lock()
			s += uint32(r.r.Int31n(1e6))
			r.mu.Unlock()
		}
		atomic.AddUint32(&BenchSink, s)
	})
}
