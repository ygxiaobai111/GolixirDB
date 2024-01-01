package util

import (
	"fmt"
	"os"
)

// checkNotExist 检查指定路径的文件或目录是否不存在
func checkNotExist(src string) bool {
	_, err := os.Stat(src)
	return os.IsNotExist(err)
}

// checkPermission 检查当前用户是否有指定路径的访问权限
func checkPermission(src string) bool {
	_, err := os.Stat(src)
	return os.IsPermission(err)
}

// isNotExistMkDir 如果指定路径不存在，则创建该路径
func isNotExistMkDir(src string) error {
	if notExist := checkNotExist(src); notExist == true {
		if err := mkDir(src); err != nil {
			return err
		}
	}
	return nil
}

// mkDir 创建指定路径的目录
func mkDir(src string) error {
	err := os.MkdirAll(src, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

// mustOpen 尝试打开（或创建）一个文件，如果必要，会先创建文件所在的目录
// fileName 是要打开的文件名，dir 是文件所在的目录
func mustOpen(fileName, dir string) (*os.File, error) {
	// 检查是否有目录的访问权限
	perm := checkPermission(dir)
	if perm == true {
		return nil, fmt.Errorf("permission denied dir: %s", dir)
	}

	// 如果目录不存在，则创建目录
	err := isNotExistMkDir(dir)
	if err != nil {
		return nil, fmt.Errorf("error during make dir %s, err: %s", dir, err)
	}

	// 打开（或创建）文件
	f, err := os.OpenFile(dir+string(os.PathSeparator)+fileName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("fail to open file, err: %s", err)
	}

	return f, nil
}
