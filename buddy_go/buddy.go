package buddy

import (
    "fmt"
    "strings"
    "sync"
    "unsafe"
)

type Buddy struct {
    rwLock       sync.RWMutex
    allocMaxSize int8
    level        int8
    pageSize     uint32
    pageCount    uint32
    recTreeSize  uint32
    recTree      []int8
    buffer       []byte
}

func IsPower2(size uint32) bool {
    value := size & (size -1)
    if value == 0 {
        return true
    }
    return false
}

func u32log2(size uint32) int8 {
    value := int8(0)
    if (size >> 16) != 0{
        size >>= 16
        value += 16
    }
    if (size >> 8) != 0 {
        size >>= 8
        value += 8
    }
    if (size >> 4) != 0 {
        size >>= 4
        value+= 4
    }
    if (size >> 2) != 0 {
        size >>= 2
        value += 2
    }
    if (size >> 1) != 0 {
        value += 1
    }
    return value
}

func FixSizePower2(size int) uint32 {
    if size <= 0 {
        return 1
    }
    newSize := uint32(size)
    if IsPower2(newSize) {
        return newSize
    }
    newSize |= newSize>>1
    newSize |= newSize>>2
    newSize |= newSize>>4
    newSize |= newSize>>8
    newSize |= newSize>>16
    return newSize + 1
}

func CreateBuddy(Cbuffer []byte, AllocMaxSize uint32, pageSize uint32) *Buddy {
    b :=new(Buddy)
    b.buffer = Cbuffer
    count := uint32(len(Cbuffer)) / pageSize
    b.level = u32log2(count)
    b.allocMaxSize = u32log2(AllocMaxSize)
    b.pageSize = pageSize
    b.pageCount = count
    b.recTreeSize = count * 2 - 1
    b.recTree = make([]int8, b.recTreeSize)
    level := b.level

    for i := uint32(0); i < b.recTreeSize; i++ {
        if IsPower2(i + 1) && i != 0 {
            level--
        }
        b.recTree[i] = level
    }
    return b
}

func (b *Buddy) Alloc(size int) (buffer []byte, err error)  {
    b.rwLock.Lock()
    defer b.rwLock.Unlock()
    index := uint32(0)
    newSize := FixSizePower2(size/int(b.pageSize))
    level := u32log2(newSize)
    if size <= 0 {
        size = int(newSize)
    }

    if b.allocMaxSize < level {
        return nil, fmt.Errorf("malloc size(%d) beyond buddy max size(%d)", size, 1 << b.allocMaxSize)
    }

    if b.recTree[0] < level {
        return nil, fmt.Errorf("buddy has no free size, size(%d) max free size(%d)",
            size, (1 << (b.recTree[0] + 1)) / 2)
    }

    for curLevel := b.level; curLevel != level; curLevel-- {
        if b.recTree[2 * index + 1] >= level{
            index = 2 * index + 1
        } else {
            index = 2 * index + 2
        }
    }
    //for curLevel := b.recTree[0]; curLevel >= level; {
    //    if 2 * index + 1 < b.recTreeSize && b.recTree[2 * index + 1] >= level{
    //        index = 2 * index + 1
    //        curLevel = b.recTree[index]
    //        continue
    //    }
    //    if 2 * index + 2 < b.recTreeSize && b.recTree[2 * index + 2] >= level {
    //        index = 2 * index + 2
    //        curLevel = b.recTree[index]
    //        continue
    //    }
    //    break
    //}

    b.recTree[index] = -1
    off := uint32( (index + 1) * (1 << level)) - b.pageCount
    maxLevel := int8(0)
    for ; index != 0; {
        index = (index + 1) /2 - 1
        maxLevel = b.recTree[2 * index + 1]
        if b.recTree[2 * index + 1] < b.recTree[2 * index + 2] {
            maxLevel = b.recTree[2 * index + 2]
        }
        b.recTree[index] = maxLevel
    }

    return b.buffer[off * b.pageSize : off * b.pageSize + uint32(size)], nil
}

func (b *Buddy) Free(buffer []byte) {
    b.rwLock.Lock()
    defer b.rwLock.Unlock()
    offLen := uintptr(unsafe.Pointer(&buffer[0])) - uintptr(unsafe.Pointer(&b.buffer[0]))
    off := uint32(offLen) / b.pageSize
    index:= off + b.pageCount - 1
    level := int8(0)
    b.recTree[index] = 0
    for ; index != 0; {
        index = (index + 1) /2 - 1
        lLevel := b.recTree[2 * index + 1]
        rLevel := b.recTree[2 * index + 2]
        if lLevel == rLevel && rLevel == level {
            b.recTree[index] = level + 1
        } else {
            if lLevel > rLevel {
                b.recTree[index] = lLevel
            } else {
                b.recTree[index] = rLevel
            }
        }
        level++
    }
    //b.Dump()
    return
}

func (b *Buddy) Dump() string{
    level := int8(0)
    count := 0
    blackCount := 1 << (b.level - level - 2)
    strBuild := strings.Builder{}

    strBuild.WriteString(fmt.Sprintf("Buddy info:\n\tpageSize:%d\n\tpageCount:%d\n\tlevel:%d\n\tmax_alloc:%d\n\tTreeArray:%v\n\tTree:",
        b.pageSize, b.pageCount, b.level, b.allocMaxSize, b.recTree))
    for i := uint32(0); i < b.recTreeSize; i++ {
        if IsPower2(i + 1) && i != 0 {
            strBuild.WriteString(fmt.Sprintf("\t\t\tlevel:%d\t count:%d\t\t\n\t", level, count))
            level++
            count = 0
            blackCount = 1 << (b.level - level)/2
        }
        count++
        for j := 0; j < blackCount; j++ {
            strBuild.WriteString("  ")
        }
        strBuild.WriteString(fmt.Sprintf("%2d", b.recTree[i]))
    }
    strBuild.WriteString(fmt.Sprintf("\t\t\tlevel:%d\t count:%d\t\t\n", level, count))
    return strBuild.String()
}

func (b *Buddy) String() string{
    return b.Dump()
}
