package wildcard

const (
	normal     = iota // 普通字符
	all               // *，匹配任意长度的字符序列
	any               // ?，匹配单个字符
	setSymbol         // []，字符集合
	rangSymbol        // [a-b]，字符范围
	negSymbol         // [^a]，否定的字符集合
)

// item 代表一个匹配单元
type item struct {
	character byte          // 匹配的字符
	set       map[byte]bool // 字符集合
	typeCode  int           // 匹配类型
}

// contains 方法用于判断给定字符是否与 item 匹配
func (i *item) contains(c byte) bool {
	// 根据不同的匹配类型来判断
	if i.typeCode == setSymbol {
		_, ok := i.set[c]
		return ok
	} else if i.typeCode == rangSymbol {
		// 范围匹配
		if _, ok := i.set[c]; ok {
			return true
		}
		// 计算范围最小和最大值
		var (
			min uint8 = 255
			max uint8 = 0
		)
		for k := range i.set {
			if min > k {
				min = k
			}
			if max < k {
				max = k
			}
		}
		return c >= min && c <= max
	} else {
		// 否定的字符集合
		_, ok := i.set[c]
		return !ok
	}
}

// Pattern 代表整个通配符模式
type Pattern struct {
	items []*item // 包含的匹配单元列表
}

// CompilePattern 将字符串形式的通配符模式转换为 Pattern 对象
func CompilePattern(src string) *Pattern {
	items := make([]*item, 0)
	escape := false // 是否转义
	inSet := false  // 是否在字符集合内
	var set map[byte]bool
	for _, v := range src {
		c := byte(v)
		if escape {
			// 处理转义字符
			items = append(items, &item{typeCode: normal, character: c})
			escape = false
		} else if c == '*' {
			items = append(items, &item{typeCode: all})
		} else if c == '?' {
			items = append(items, &item{typeCode: any})
		} else if c == '\\' {
			escape = true
		} else if c == '[' {
			// 开始字符集合
			if !inSet {
				inSet = true
				set = make(map[byte]bool)
			} else {
				set[c] = true
			}
		} else if c == ']' {
			// 结束字符集合
			if inSet {
				inSet = false
				typeCode := setSymbol
				if _, ok := set['-']; ok {
					typeCode = rangSymbol
					delete(set, '-')
				}
				if _, ok := set['^']; ok {
					typeCode = negSymbol
					delete(set, '^')
				}
				items = append(items, &item{typeCode: typeCode, set: set})
			} else {
				items = append(items, &item{typeCode: normal, character: c})
			}
		} else {
			// 普通字符或字符集合内的字符
			if inSet {
				set[c] = true
			} else {
				items = append(items, &item{typeCode: normal, character: c})
			}
		}
	}
	return &Pattern{
		items: items,
	}
}

// IsMatch 方法判断字符串是否与 Pattern 匹配
func (p *Pattern) IsMatch(s string) bool {
	if len(p.items) == 0 {
		return len(s) == 0
	}
	m := len(s)
	n := len(p.items)
	// 动态规划表
	table := make([][]bool, m+1)
	for i := 0; i < m+1; i++ {
		table[i] = make([]bool, n+1)
	}
	// 初始化动态规划表
	table[0][0] = true
	for j := 1; j < n+1; j++ {
		table[0][j] = table[0][j-1] && p.items[j-1].typeCode == all
	}
	for i := 1; i < m+1; i++ {
		for j := 1; j < n+1; j++ {
			// 根据不同的匹配类型更新动态规划表
			if p.items[j-1].typeCode == all {
				table[i][j] = table[i-1][j] || table[i][j-1]
			} else {
				table[i][j] = table[i-1][j-1] &&
					(p.items[j-1].typeCode == any ||
						(p.items[j-1].typeCode == normal && uint8(s[i-1]) == p.items[j-1].character) ||
						(p.items[j-1].typeCode >= setSymbol && p.items[j-1].contains(s[i-1])))
			}
		}
	}
	return table[m][n]
}
