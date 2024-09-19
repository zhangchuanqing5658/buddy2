package buddy

import (
    "container/list"
    "fmt"
    "math/rand"
    "sync"
    "testing"
)

const (
    LoopCount = 20
    BuffSize  = 64
    MaxAlloc  = 16
)

func initRecord (buff []byte, buffPtr [][]byte) {
    for i := 0; i < len(buff); i++ {
        buff[i] = 0
        buffPtr[i] = nil
    }
}

func TestBuddy_Loop(t *testing.T) {
    buff := make([]byte, BuffSize)
    allocRec := make([][]byte, BuffSize)
    b := CreateBuddy(buff, 16, 1)
    fmt.Printf("%s\n", b.Dump())
    for i := 0; i < LoopCount; i++ {
        initRecord(buff, allocRec)
        index := 0
        mallocTot := 0
        for ;b.recTree[0] >= 0; {
            size := rand.Int() % MaxAlloc
            allocBuff, err := b.Alloc(size)
            if err != nil {
                continue
            }
            for j := 0; j < len(allocBuff); j++ {
                if allocBuff[j] != 0 {
                    goto replayFailed
                }
                allocBuff[j] = 1
            }
            allocRec[index] = allocBuff
            mallocTot += len(allocBuff)
            index++
            //b.Dump()
        }
        fmt.Printf("\n\n\n*****************\n")
        fmt.Printf("loop:%d\n\tmalloc %d times: [%d]; max free:%d\n", i + 1, index, mallocTot, b.recTree[0])
        fmt.Printf("%s\n", b.Dump())
        fmt.Printf("\tmem bits:");
        for j := 0; j < BuffSize; j++ {
            if j % 4 == 0 {
                fmt.Printf(" ")
            }
            fmt.Printf("%d", buff[j]);
        }

        fmt.Printf("\n\n\nenter free");
        for j := 0; j < BuffSize; j++ {
            if allocRec[j] == nil {
                break
            }
            b.Free(allocRec[j])
        }

        if b.recTree[0] != b.level {
            fmt.Printf("%s\n", b.Dump())
            goto replayFailed
        }

        fmt.Printf("\n\n\tfree loop:%d\n", index)
        fmt.Printf("%s\n", b.Dump())
        fmt.Printf("\n*****************\n\n\n")
    }

    return
replayFailed:

    fmt.Printf("\nfailed test\n")
    return
}

type MFreeList struct {
    free list.List
    rwLock sync.RWMutex
}

func CreateMlist() *MFreeList {
    m := new(MFreeList)
    for i := 0; i < 1024; i++ {
        m.free.PushBack(make([]byte, 1 * 1024))
    }

    return m
}

func (m *MFreeList) Alloc() []byte {
    m.rwLock.Lock()
    defer m.rwLock.Unlock()
    elem := m.free.Front()
    m.free.Remove(elem)
    return elem.Value.([]byte)
}

func (m *MFreeList) Free(buffer []byte) {
    m.rwLock.Lock()
    defer m.rwLock.Unlock()
    m.free.PushBack(buffer)
    return
}

func BenchmarkFreeList_Loop(b *testing.B) {
    m := CreateMlist()
    wg := sync.WaitGroup{}

    b.ResetTimer()
    for n := 0; n < b.N; n++ {
        for i := 0; i < 4; i++ {
            wg.Add(1)
            go func() {
                defer wg.Done()
                for j := 0; j < 1024; j++ {
                    buffer := m.Alloc()
                    m.Free(buffer)
                }
            }()
        }
        wg.Wait()
    }

    b.ReportAllocs()
}

const PoolCnt = 9

func BenchmarkSyncPool_Loop(b *testing.B) {

    var PoolArr [PoolCnt]*sync.Pool
    INITKB := 1 << 12
    for i := 0; i < PoolCnt; i++ {
        PoolArr[i] = &sync.Pool{New: func() interface{} {
            return make([]byte, INITKB << i)
        }}
    }

    wg := sync.WaitGroup{}
    b.ResetTimer()
    for n := 0; n < b.N; n++ {
        for i := 0; i < 4; i++ {
            wg.Add(1)
            go func() {
                defer wg.Done()
                for j := 0; j < 1024; j++ {
                    size := rand.Int() % (1024 / 4) + 1
                    index := u32log2(uint32(size))
                    buffer := PoolArr[index].Get()
                    PoolArr[index].Put(buffer)
                }
            }()
        }
        wg.Wait()
    }

    b.ReportAllocs()
}

func BenchmarkBuddy_Loop(b *testing.B) {

    allocator := CreateBuddy(make([]byte, 4 * 1024 * 1024), 1 * 1024 * 1024, 4 * 1024)
    wg := sync.WaitGroup{}
    b.ResetTimer()
    for n := 0; n < b.N; n++ {
        for i := 0; i < 4; i++ {
            wg.Add(1)
            go func() {
                defer wg.Done()
                for j := 0; j < 1024; j++ {
                    size := rand.Int() % (1024 / 4) + 1
                    buffer, err := allocator.Alloc(size << 12)
                    if err != nil {
                        b.Logf("malloc size:%d failed:%s", size, err.Error())
                        continue
                    }
                    allocator.Free(buffer)
                }
            }()
        }
        wg.Wait()
    }
    b.ReportAllocs()
}

func BenchmarkParallelFreeList(b *testing.B) {
    m := CreateMlist()
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            buffer := m.Alloc()
            m.Free(buffer)
        }
    })
}

func BenchmarkParallelSyncPool(b *testing.B) {
    var PoolArr [PoolCnt]*sync.Pool
    INITKB := 1 << 12
    for i := 0; i < PoolCnt; i++ {
        PoolArr[i] = &sync.Pool{New: func() interface{} {
            return make([]byte, INITKB << i)
        }}
    }

    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            size := rand.Int() % (1024 / 4) + 1
            index := u32log2(uint32(size))
            buffer := PoolArr[index].Get()
            PoolArr[index].Put(buffer)
        }
    })
}


func BenchmarkParallelBuddy(b *testing.B) {
    allocator := CreateBuddy(make([]byte, 16 * 1024 * 1024), 1 * 1024 * 1024, 4 * 1024)
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            size := rand.Int() % (1024 / 4) + 1
            buffer, err := allocator.Alloc(size << 12)
            if err != nil {
                b.Logf("malloc size:%d failed:%s", size, err.Error())
                continue
            }
            allocator.Free(buffer)
        }
    })
}
