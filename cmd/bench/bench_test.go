package main

import (
	"testing"
	"unsafe"
)

const (
	testSize   = 1000
	iterations = 1000
)

// Structure of Arrays (SoA) - separate slices (typical approach).
type SeparateSlices struct {
	countedWords []uint16
	totalWords   []uint16
	bookIds      []uint16
}

// Array of Structures (AoS) - traditional approach.
type BookEntry struct {
	countedWords uint16
	totalWords   uint16
	bookId       uint16
}

type ArrayOfStructs struct {
	entries []BookEntry
}

// Optimized: Single allocation with slice views.
type OptimizedLayout struct {
	backing []uint16 // Single allocation

	// Slice views into the backing array
	countedWords []uint16
	totalWords   []uint16
	bookIds      []uint16
}

// Cache-aligned layout with padding.
type CacheAlignedLayout struct {
	backing []uint16

	countedWords []uint16
	totalWords   []uint16
	bookIds      []uint16

	// Padding to ensure each slice starts at cache boundary
	_ [32]uint16 // Padding between arrays
}

// Initialize separate slices (poor cache behavior).
func NewSeparateSlices(size int) *SeparateSlices {
	return &SeparateSlices{
		countedWords: make([]uint16, size),
		totalWords:   make([]uint16, size),
		bookIds:      make([]uint16, size),
	}
}

// Initialize array of structs (good cache behavior for small structs).
func NewArrayOfStructs(size int) *ArrayOfStructs {
	return &ArrayOfStructs{
		entries: make([]BookEntry, size),
	}
}

// Initialize optimized layout (excellent cache behavior).
func NewOptimizedLayout(size int) *OptimizedLayout {
	// Allocate single backing array for all data
	backing := make([]uint16, size*3) // 3 fields per entry

	return &OptimizedLayout{
		backing:      backing,
		countedWords: backing[0:size],          // First third
		totalWords:   backing[size : size*2],   // Second third
		bookIds:      backing[size*2 : size*3], // Final third
	}
}

// Initialize cache-aligned layout.
func NewCacheAlignedLayout(size int) *CacheAlignedLayout {
	// Calculate cache-aligned sizes
	const cacheLineSize = 64
	const uint16Size = 2
	elementsPerLine := cacheLineSize / uint16Size // 32 elements

	// Round each section to cache line boundary
	alignedSize := ((size + elementsPerLine - 1) / elementsPerLine) * elementsPerLine

	backing := make([]uint16, alignedSize*3)

	return &CacheAlignedLayout{
		backing:      backing,
		countedWords: backing[0:size],
		totalWords:   backing[alignedSize : alignedSize+size],
		bookIds:      backing[alignedSize*2 : alignedSize*2+size],
	}
}

// Benchmark separate slices - poor cache performance.
func BenchmarkSeparateSlices(b *testing.B) {
	ss := NewSeparateSlices(testSize)

	// Initialize with test data
	for i := 0; i < testSize; i++ {
		ss.countedWords[i] = uint16(i)
		ss.totalWords[i] = uint16(i * 2)
		ss.bookIds[i] = uint16(i * 3)
	}

	b.ResetTimer()
	b.ReportAllocs()

	var sum uint16
	for i := 0; i < b.N; i++ {
		for j := 0; j < testSize; j++ {
			// Access pattern: a[i], b[i], c[i] - likely 3 cache misses
			sum += ss.countedWords[j] + ss.totalWords[j] + ss.bookIds[j]
		}
	}
	_ = sum
}

// Benchmark array of structs - good cache locality for small structs.
func BenchmarkArrayOfStructs(b *testing.B) {
	aos := NewArrayOfStructs(testSize)

	for i := 0; i < testSize; i++ {
		aos.entries[i] = BookEntry{
			countedWords: uint16(i),
			totalWords:   uint16(i * 2),
			bookId:       uint16(i * 3),
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	var sum uint16
	for i := 0; i < b.N; i++ {
		for j := 0; j < testSize; j++ {
			// All fields in same cache line - 1 cache miss for ~10 entries
			entry := &aos.entries[j]
			sum += entry.countedWords + entry.totalWords + entry.bookId
		}
	}
	_ = sum
}

// Benchmark optimized layout - excellent cache behavior.
func BenchmarkOptimizedLayout(b *testing.B) {
	ol := NewOptimizedLayout(testSize)

	for i := 0; i < testSize; i++ {
		ol.countedWords[i] = uint16(i)
		ol.totalWords[i] = uint16(i * 2)
		ol.bookIds[i] = uint16(i * 3)
	}

	b.ResetTimer()
	b.ReportAllocs()

	var sum uint16
	for i := 0; i < b.N; i++ {
		for j := 0; j < testSize; j++ {
			// Better cache behavior - slices are closer in memory
			sum += ol.countedWords[j] + ol.totalWords[j] + ol.bookIds[j]
		}
	}
	_ = sum
}

// Benchmark cache-aligned layout.
func BenchmarkCacheAlignedLayout(b *testing.B) {
	cal := NewCacheAlignedLayout(testSize)

	for i := 0; i < testSize; i++ {
		cal.countedWords[i] = uint16(i)
		cal.totalWords[i] = uint16(i * 2)
		cal.bookIds[i] = uint16(i * 3)
	}

	b.ResetTimer()
	b.ReportAllocs()

	var sum uint16
	for i := 0; i < b.N; i++ {
		for j := 0; j < testSize; j++ {
			sum += cal.countedWords[j] + cal.totalWords[j] + cal.bookIds[j]
		}
	}
	_ = sum
}

// Benchmark different access patterns.
func BenchmarkSequentialVsStrided(b *testing.B) {
	ol := NewOptimizedLayout(testSize)

	for i := 0; i < testSize; i++ {
		ol.countedWords[i] = uint16(i)
		ol.totalWords[i] = uint16(i * 2)
		ol.bookIds[i] = uint16(i * 3)
	}

	b.Run("Sequential", func(b *testing.B) {
		var sum uint16
		for i := 0; i < b.N; i++ {
			// Access each array completely before moving to next
			for j := 0; j < testSize; j++ {
				sum += ol.countedWords[j]
			}
			for j := 0; j < testSize; j++ {
				sum += ol.totalWords[j]
			}
			for j := 0; j < testSize; j++ {
				sum += ol.bookIds[j]
			}
		}
		_ = sum
	})

	b.Run("Strided", func(b *testing.B) {
		var sum uint16
		for i := 0; i < b.N; i++ {
			// Access a[i], b[i], c[i] pattern
			for j := 0; j < testSize; j++ {
				sum += ol.countedWords[j] + ol.totalWords[j] + ol.bookIds[j]
			}
		}
		_ = sum
	})
}

// Memory layout analysis.
func BenchmarkMemoryLayout(b *testing.B) {
	b.Run("Analyze_Separate_Slices", func(b *testing.B) {
		ss := NewSeparateSlices(10)

		p1 := unsafe.Pointer(&ss.countedWords[0])
		p2 := unsafe.Pointer(&ss.totalWords[0])
		p3 := unsafe.Pointer(&ss.bookIds[0])

		dist12 := uintptr(p2) - uintptr(p1)
		dist23 := uintptr(p3) - uintptr(p2)

		b.ReportMetric(float64(dist12), "bytes_between_slice1_slice2")
		b.ReportMetric(float64(dist23), "bytes_between_slice2_slice3")
		b.ReportMetric(float64(dist12/64), "cache_lines_between_slice1_slice2")
	})

	b.Run("Analyze_Optimized_Layout", func(b *testing.B) {
		ol := NewOptimizedLayout(10)

		p1 := unsafe.Pointer(&ol.countedWords[0])
		p2 := unsafe.Pointer(&ol.totalWords[0])
		p3 := unsafe.Pointer(&ol.bookIds[0])

		dist12 := uintptr(p2) - uintptr(p1)
		dist23 := uintptr(p3) - uintptr(p2)

		b.ReportMetric(float64(dist12), "bytes_between_slice1_slice2")
		b.ReportMetric(float64(dist23), "bytes_between_slice2_slice3")
		b.ReportMetric(float64(dist12/64), "cache_lines_between_slice1_slice2")
	})
}

// Cache miss simulation.
func BenchmarkCacheMissSimulation(b *testing.B) {
	// Simulate worst case: slices far apart in memory
	const separation = 1024 * 1024 // 1MB apart to ensure cache misses

	backing := make([]uint16, testSize*3+separation*2/2) // uint16 = 2 bytes

	slice1 := backing[0:testSize]
	slice2 := backing[testSize+separation/2 : testSize*2+separation/2]
	slice3 := backing[testSize*2+separation : testSize*3+separation]

	b.ResetTimer()
	var sum uint16
	for i := 0; i < b.N; i++ {
		for j := 0; j < testSize; j++ {
			// Guaranteed cache misses due to large separation
			sum += slice1[j] + slice2[j] + slice3[j]
		}
	}
	_ = sum
}

func main() {
	// Run: go test -bench=. -benchmem
	// Compare the different memory layout approaches
}
