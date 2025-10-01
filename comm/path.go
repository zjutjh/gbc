package comm

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

func ParseKey(key, pkgName, prefix, suffix string) (string, string, string, error) {
	ks := strings.Split(key, ".")
	path := strings.Join(ks, "/")
	name := ks[len(ks)-1]
	if name == "" {
		return "", "", "", fmt.Errorf("key不能以[.]结尾")
	}
	ns := strings.Split(name, "_")
	for k, v := range ns {
		ns[k] = strings.ToUpper(v[:1]) + v[1:]
	}
	structName := strings.Join(ns, "")
	packageName := pkgName
	if len(ks) > 1 {
		packageName = ks[len(ks)-2]
		dir := prefix + strings.Join(ks[:len(ks)-1], "/")
		fi, err := os.Stat(dir)
		if os.IsNotExist(err) {
			// 需要创建目录
			os.Mkdir(dir, 0755)
		} else if !fi.IsDir() {
			return "", "", "", errors.New("路径[" + dir + "]已存在且不是一个目录")
		}
	}
	return prefix + path + suffix, structName, packageName, nil
}
