package service

import (
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	cryptorand "crypto/rand"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"io"
	"math"
	"math/big"
	"math/rand/v2"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/leonelquinteros/gotext"

	"github.com/tnb-labs/panel/internal/http/request"
)

type ToolboxBenchmarkService struct {
	t *gotext.Locale
}

func NewToolboxBenchmarkService(t *gotext.Locale) *ToolboxBenchmarkService {
	return &ToolboxBenchmarkService{
		t: t,
	}
}

// Test 运行测试
func (s *ToolboxBenchmarkService) Test(w http.ResponseWriter, r *http.Request) {
	req, err := Bind[request.ToolboxBenchmarkTest](r)
	if err != nil {
		Error(w, http.StatusUnprocessableEntity, "%v", err)
		return
	}

	switch req.Name {
	case "image":
		result := s.imageProcessing(req.Multi)
		Success(w, result)
	case "machine":
		result := s.machineLearning(req.Multi)
		Success(w, result)
	case "compile":
		result := s.compileSimulationSingle(req.Multi)
		Success(w, result)
	case "encryption":
		result := s.encryptionTest(req.Multi)
		Success(w, result)
	case "compression":
		result := s.compressionTest(req.Multi)
		Success(w, result)
	case "physics":
		result := s.physicsSimulation(req.Multi)
		Success(w, result)
	case "json":
		result := s.jsonProcessing(req.Multi)
		Success(w, result)
	case "disk":
		result := s.diskTestTask()
		Success(w, result)
	case "memory":
		result := s.memoryTestTask()
		Success(w, result)
	default:
		Error(w, http.StatusUnprocessableEntity, s.t.Get("unknown test type"))
	}
}

// calculateCpuScore 计算CPU成绩
func (s *ToolboxBenchmarkService) calculateCpuScore(duration time.Duration) int {
	score := int((10 / duration.Seconds()) * float64(3000))

	if score < 0 {
		score = 0
	}
	return score
}

// calculateScore 计算内存/硬盘成绩
func (s *ToolboxBenchmarkService) calculateScore(duration time.Duration) int {
	score := int((20 / duration.Seconds()) * float64(30000))

	if score < 0 {
		score = 0
	}
	return score
}

// 图像处理

func (s *ToolboxBenchmarkService) imageProcessing(multi bool) int {
	n := 1
	if multi {
		n = runtime.NumCPU()
	}
	start := time.Now()
	if err := s.imageProcessingTask(n); err != nil {
		return 0
	}
	duration := time.Since(start)
	return s.calculateCpuScore(duration)
}

func (s *ToolboxBenchmarkService) imageProcessingTask(numThreads int) error {
	img := image.NewRGBA(image.Rect(0, 0, 4000, 4000))
	for x := 0; x < 4000; x++ {
		for y := 0; y < 4000; y++ {
			img.Set(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), A: 255})
		}
	}

	var wg sync.WaitGroup
	dx := img.Bounds().Dx()
	dy := img.Bounds().Dy()
	chunkSize := dy / numThreads

	for i := 0; i < numThreads; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			startY := i * chunkSize
			endY := startY + chunkSize
			if i == numThreads-1 {
				endY = dy
			}
			for x := 1; x < dx-1; x++ {
				for y := startY + 1; y < endY-1; y++ {
					// 卷积操作（模糊）
					rTotal, gTotal, bTotal := 0, 0, 0
					for k := -1; k <= 1; k++ {
						for l := -1; l <= 1; l++ {
							r, g, b, _ := img.At(x+k, y+l).RGBA()
							rTotal += int(r)
							gTotal += int(g)
							bTotal += int(b)
						}
					}
					rAvg := uint8(rTotal / 9 / 256)
					gAvg := uint8(gTotal / 9 / 256)
					bAvg := uint8(bTotal / 9 / 256)
					img.Set(x, y, color.RGBA{R: rAvg, G: gAvg, B: bAvg, A: 255})
				}
			}
		}(i)
	}

	wg.Wait()
	return nil
}

// 机器学习（矩阵乘法）

func (s *ToolboxBenchmarkService) machineLearning(multi bool) int {
	n := 1
	if multi {
		n = runtime.NumCPU()
	}
	start := time.Now()
	s.machineLearningTask(n)
	duration := time.Since(start)
	return s.calculateCpuScore(duration)
}

func (s *ToolboxBenchmarkService) machineLearningTask(numThreads int) {
	size := 900
	a := make([][]float64, size)
	b := make([][]float64, size)
	for i := 0; i < size; i++ {
		a[i] = make([]float64, size)
		b[i] = make([]float64, size)
		for j := 0; j < size; j++ {
			a[i][j] = rand.Float64()
			b[i][j] = rand.Float64()
		}
	}

	c := make([][]float64, size)
	for i := 0; i < size; i++ {
		c[i] = make([]float64, size)
	}

	var wg sync.WaitGroup
	chunkSize := size / numThreads

	for k := 0; k < numThreads; k++ {
		wg.Add(1)
		go func(k int) {
			defer wg.Done()
			start := k * chunkSize
			end := start + chunkSize
			if k == numThreads-1 {
				end = size
			}
			for i := start; i < end; i++ {
				for j := 0; j < size; j++ {
					sum := 0.0
					for l := 0; l < size; l++ {
						sum += a[i][l] * b[l][j]
					}
					c[i][j] = sum
				}
			}
		}(k)
	}

	wg.Wait()
}

// 数学问题（计算斐波那契数）

func (s *ToolboxBenchmarkService) compileSimulationSingle(multi bool) int {
	n := 1
	if multi {
		n = runtime.NumCPU()
	}
	start := time.Now()
	totalCalculations := 1000
	fibNumber := 20000

	calculationsPerThread := totalCalculations / n
	remainder := totalCalculations % n

	var wg sync.WaitGroup

	for i := 0; i < n; i++ {
		tasks := calculationsPerThread
		if i < remainder {
			tasks++ // 处理无法均分的剩余任务
		}
		wg.Add(1)
		go func(tasks int) {
			defer wg.Done()
			for j := 0; j < tasks; j++ {
				s.fib(fibNumber)
			}
		}(tasks)
	}

	wg.Wait()
	duration := time.Since(start)
	return s.calculateCpuScore(duration)
}

// 斐波那契函数
func (s *ToolboxBenchmarkService) fib(n int) *big.Int {
	if n < 2 {
		return big.NewInt(int64(n))
	}
	a := big.NewInt(0)
	b := big.NewInt(1)
	temp := big.NewInt(0)
	for i := 2; i <= n; i++ {
		temp.Add(a, b)
		a.Set(b)
		b.Set(temp)
	}
	return b
}

// AES加密

func (s *ToolboxBenchmarkService) encryptionTest(multi bool) int {
	n := 1
	if multi {
		n = runtime.NumCPU()
	}
	start := time.Now()
	if err := s.encryptionTestTask(n); err != nil {
		return 0
	}
	duration := time.Since(start)
	return s.calculateCpuScore(duration)
}

func (s *ToolboxBenchmarkService) encryptionTestTask(numThreads int) error {
	key := []byte("abcdefghijklmnopqrstuvwxyz123456")
	dataSize := 1024 * 1024 * 512 // 512 MB
	plaintext := []byte(strings.Repeat("A", dataSize))
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	chunkSize := dataSize / numThreads

	var wg sync.WaitGroup

	for i := 0; i < numThreads; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			start := i * chunkSize
			end := start + chunkSize
			if i == numThreads-1 {
				end = dataSize
			}

			nonce := make([]byte, aesGCM.NonceSize())
			if _, err = cryptorand.Read(nonce); err != nil {
				return
			}

			aesGCM.Seal(nil, nonce, plaintext[start:end], nil)
		}(i)
	}

	wg.Wait()
	return nil
}

// 压缩/解压缩

func (s *ToolboxBenchmarkService) compressionTest(multi bool) int {
	n := 1
	if multi {
		n = runtime.NumCPU()
	}
	start := time.Now()
	s.compressionTestTask(n)
	duration := time.Since(start)
	return s.calculateCpuScore(duration)
}

func (s *ToolboxBenchmarkService) compressionTestTask(numThreads int) {
	data := []byte(strings.Repeat("耗子面板", 50000000))
	chunkSize := len(data) / numThreads

	var wg sync.WaitGroup

	compressedChunks := make([]bytes.Buffer, numThreads)

	// 压缩
	for i := 0; i < numThreads; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			start := i * chunkSize
			end := start + chunkSize
			if i == numThreads-1 {
				end = len(data)
			}
			var buf bytes.Buffer
			w := gzip.NewWriter(&buf)
			_, _ = w.Write(data[start:end])
			_ = w.Close()
			compressedChunks[i] = buf
		}(i)
	}

	wg.Wait()

	// 解压缩
	for i := 0; i < numThreads; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			r, err := gzip.NewReader(&compressedChunks[i])
			if err != nil {
				return
			}
			_, err = io.Copy(io.Discard, r)
			if err != nil {
				return
			}
			_ = r.Close()
		}(i)
	}

	wg.Wait()
}

// 物理仿真（N体问题）

func (s *ToolboxBenchmarkService) physicsSimulation(multi bool) int {
	n := 1
	if multi {
		n = runtime.NumCPU()
	}
	start := time.Now()
	s.physicsSimulationTask(n)
	duration := time.Since(start)
	return s.calculateCpuScore(duration)
}

func (s *ToolboxBenchmarkService) physicsSimulationTask(numThreads int) {
	const (
		numBodies = 4000
		steps     = 30
	)

	type Body struct {
		x, y, z, vx, vy, vz float64
	}

	bodies := make([]Body, numBodies)
	for i := 0; i < numBodies; i++ {
		bodies[i] = Body{
			x:  rand.Float64(),
			y:  rand.Float64(),
			z:  rand.Float64(),
			vx: rand.Float64(),
			vy: rand.Float64(),
			vz: rand.Float64(),
		}
	}

	chunkSize := numBodies / numThreads

	for step := 0; step < steps; step++ {
		var wg sync.WaitGroup

		// 更新速度
		for k := 0; k < numThreads; k++ {
			wg.Add(1)
			go func(k int) {
				defer wg.Done()
				start := k * chunkSize
				end := start + chunkSize
				if k == numThreads-1 {
					end = numBodies
				}
				for i := start; i < end; i++ {
					bi := &bodies[i]
					for j := 0; j < numBodies; j++ {
						if i == j {
							continue
						}
						bj := &bodies[j]
						dx := bj.x - bi.x
						dy := bj.y - bi.y
						dz := bj.z - bi.z
						dist := math.Sqrt(dx*dx + dy*dy + dz*dz)
						if dist == 0 {
							continue
						}
						force := 1 / (dist * dist)
						bi.vx += force * dx / dist
						bi.vy += force * dy / dist
						bi.vz += force * dz / dist
					}
				}
			}(k)
		}

		wg.Wait()

		// 更新位置
		for k := 0; k < numThreads; k++ {
			wg.Add(1)
			go func(k int) {
				defer wg.Done()
				start := k * chunkSize
				end := start + chunkSize
				if k == numThreads-1 {
					end = numBodies
				}
				for i := start; i < end; i++ {
					bi := &bodies[i]
					bi.x += bi.vx
					bi.y += bi.vy
					bi.z += bi.vz
				}
			}(k)
		}

		wg.Wait()
	}
}

// JSON解析

func (s *ToolboxBenchmarkService) jsonProcessing(multi bool) int {
	n := 1
	if multi {
		n = runtime.NumCPU()
	}
	start := time.Now()
	s.jsonProcessingTask(n)
	duration := time.Since(start)
	return s.calculateCpuScore(duration)
}

func (s *ToolboxBenchmarkService) jsonProcessingTask(numThreads int) {
	numElements := 1000000
	elementsPerThread := numElements / numThreads

	var wg sync.WaitGroup
	for i := 0; i < numThreads; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			start := i * elementsPerThread
			end := start + elementsPerThread
			if i == numThreads-1 {
				end = numElements
			}

			elements := make([]map[string]any, 0, end-start)
			for j := start; j < end; j++ {
				elements = append(elements, map[string]any{
					"id":    j,
					"value": fmt.Sprintf("Value%d", j),
				})
			}
			encoded, err := json.Marshal(elements)
			if err != nil {
				return
			}

			var parsed []map[string]any
			err = json.Unmarshal(encoded, &parsed)
			if err != nil {
				return
			}
		}(i)
	}

	wg.Wait()
}

// 内存性能

func (s *ToolboxBenchmarkService) memoryTestTask() map[string]any {
	results := make(map[string]any)
	dataSize := 500 * 1024 * 1024 // 500 MB
	data := make([]byte, dataSize)
	_, _ = cryptorand.Read(data)

	start := time.Now()
	// 内存读写速度
	results["bandwidth"] = s.memoryBandwidthTest(data)
	// 内存访问延迟
	data = data[:100*1024*1024] // 100 MB
	results["latency"] = s.memoryLatencyTest(data)
	duration := time.Since(start)
	results["score"] = s.calculateScore(duration)

	return results
}

func (s *ToolboxBenchmarkService) memoryBandwidthTest(data []byte) string {
	dataSize := len(data)

	startTime := time.Now()

	for i := 0; i < dataSize; i++ {
		data[i] ^= 0xFF
	}

	duration := time.Since(startTime).Seconds()
	if duration == 0 {
		return "N/A"
	}
	speed := float64(dataSize) / duration / (1024 * 1024)
	return fmt.Sprintf("%.2f MB/s", speed)
}

func (s *ToolboxBenchmarkService) memoryLatencyTest(data []byte) string {
	dataSize := len(data)
	indices := rand.Perm(dataSize)

	startTime := time.Now()
	sum := byte(0)
	for _, idx := range indices {
		sum ^= data[idx]
	}
	duration := time.Since(startTime).Seconds()
	if duration == 0 {
		return "N/A"
	}
	avgLatency := duration * 1e9 / float64(dataSize)
	return fmt.Sprintf("%.2f ns", avgLatency)
}

// 硬盘IO

func (s *ToolboxBenchmarkService) diskTestTask() map[string]any {
	results := make(map[string]any)
	blockSizes := []int64{4 * 1024, 64 * 1024, 512 * 1024, 1 * 1024 * 1024} // 4K, 64K, 512K, 1M
	fileSize := int64(100 * 1024 * 1024)                                    // 100MB 文件

	start := time.Now()
	for _, blockSize := range blockSizes {
		result := s.diskIOTest(blockSize, fileSize)
		results[fmt.Sprintf("%d", blockSize/1024)] = result
	}
	duration := time.Since(start)
	results["score"] = s.calculateScore(duration)

	return results
}

func (s *ToolboxBenchmarkService) diskIOTest(blockSize int64, fileSize int64) map[string]any {
	result := make(map[string]any)
	tempFile := fmt.Sprintf("tempfile_%d", blockSize)
	defer func(name string) {
		_ = os.Remove(name)
	}(tempFile)

	// 写测试
	writeSpeed, writeIOPS := s.diskWriteTest(tempFile, blockSize, fileSize)
	// 读测试
	readSpeed, readIOPS := s.diskReadTest(tempFile, blockSize, fileSize)

	result["write_speed"] = fmt.Sprintf("%.2f MB/s", writeSpeed)
	result["write_iops"] = fmt.Sprintf("%.2f IOPS", writeIOPS)
	result["read_speed"] = fmt.Sprintf("%.2f MB/s", readSpeed)
	result["read_iops"] = fmt.Sprintf("%.2f IOPS", readIOPS)

	return result
}

func (s *ToolboxBenchmarkService) diskWriteTest(fileName string, blockSize int64, fileSize int64) (float64, float64) {
	totalBlocks := fileSize / blockSize

	data := make([]byte, blockSize)
	_, _ = cryptorand.Read(data)

	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_SYNC, 0644)
	if err != nil {
		return 0, 0
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	start := time.Now()

	for i := int64(0); i < totalBlocks; i++ {
		// 生成随机偏移
		offset := rand.Int64N(fileSize - blockSize + 1)
		_, err := file.WriteAt(data, offset)
		if err != nil {
			return 0, 0
		}
	}

	_ = file.Sync()

	duration := time.Since(start).Seconds()
	if duration == 0 {
		duration = 1
	}
	speed := float64(totalBlocks*blockSize) / duration / (1024 * 1024)
	iops := float64(totalBlocks) / duration
	return speed, iops
}

func (s *ToolboxBenchmarkService) diskReadTest(fileName string, blockSize int64, fileSize int64) (float64, float64) {
	totalBlocks := fileSize / blockSize

	data := make([]byte, blockSize)

	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_SYNC, 0644)
	if err != nil {
		return 0, 0
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	start := time.Now()

	for i := int64(0); i < totalBlocks; i++ {
		// 生成随机偏移
		offset := rand.Int64N(fileSize - blockSize + 1)
		_, err := file.ReadAt(data, offset)
		if err != nil && err != io.EOF {
			return 0, 0
		}
	}

	duration := time.Since(start).Seconds()
	if duration == 0 {
		duration = 1
	}
	speed := float64(totalBlocks*blockSize) / duration / (1024 * 1024)
	iops := float64(totalBlocks) / duration
	return speed, iops
}
