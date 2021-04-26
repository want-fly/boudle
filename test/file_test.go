package test

import (
	"boudle/until"
	"fmt"
	"testing"
)

// TestFileDeCompress 测试解压tar包 执行完之后使用 diff -urNa 来比较我们解压和 tar -zxvf   的文件是否相同
func TestFileDeCompress(t *testing.T) {
	err := until.DeCompress("/root/remote-project-file/wordpress-5.7.tar.gz", "/root/remote-project-file/test/")
	if err != nil {
		fmt.Println("err:", err)
		return
	}
}
