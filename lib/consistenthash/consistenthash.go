package consistenthash

import (
	"hash/crc32"
	"sort"
)

// HashFunc 定义了生成哈希码的函数类型
type HashFunc func(data []byte) uint32

// NodeMap 存储节点，并且可以从 NodeMap 中选择节点
type NodeMap struct {
	hashFunc    HashFunc       // 用于生成哈希值的函数
	nodeHashs   []int          // 存储节点哈希值的已排序切片
	nodehashMap map[int]string // 哈希值到节点名称的映射
}

// NewNodeMap 创建一个新的 NodeMap
func NewNodeMap(fn HashFunc) *NodeMap {
	m := &NodeMap{
		hashFunc:    fn,
		nodehashMap: make(map[int]string),
	}
	if m.hashFunc == nil {
		m.hashFunc = crc32.ChecksumIEEE // 默认使用 crc32.ChecksumIEEE 作为哈希函数
	}
	return m
}

// IsEmpty 返回 NodeMap 是否为空（没有节点）
func (m *NodeMap) IsEmpty() bool {
	return len(m.nodeHashs) == 0
}

// AddNode 将给定的节点添加到一致性哈希环中
func (m *NodeMap) AddNode(keys ...string) {
	for _, key := range keys {
		if key == "" {
			continue
		}
		hash := int(m.hashFunc([]byte(key)))
		m.nodeHashs = append(m.nodeHashs, hash)
		m.nodehashMap[hash] = key
	}
	sort.Ints(m.nodeHashs) // 对节点哈希值进行排序
}

// PickNode 根据提供的键值，在哈希环上获取最接近的节点
func (m *NodeMap) PickNode(key string) string {
	if m.IsEmpty() {
		return ""
	}

	hash := int(m.hashFunc([]byte(key)))

	// 通过二分查找找到适当的副本
	idx := sort.Search(len(m.nodeHashs), func(i int) bool {
		return m.nodeHashs[i] >= hash
	})

	// 如果 idx 等于节点哈希值的长度，意味着已经环绕到了第一个副本
	if idx == len(m.nodeHashs) {
		idx = 0
	}

	return m.nodehashMap[m.nodeHashs[idx]]
}
