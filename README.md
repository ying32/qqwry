# qqwry
纯真ip数据库查询

本代码改自[原自己的Delphi版本](http://blog.csdn.net/zyjying520/article/details/8373931 "原自己的Delphi版本")

by:ying32   2016/5/13

未测试多种带有重定义或未重定义的IP，如果查询出错，可发邮件给我，告知改进，或者由您改进，但记得发我一份。

[纯真网络](http://www.cz88.net/ "纯真网络")
[纯真IP数据格式](http://lumaqq.linuxsir.org/article/qqwry_format_detail.html "纯真IP数据格式")

#### golang使用例程
```go
import (
	"fmt"
	"github.com/ying32/qqwry"
)

func main() {
	qdat := qqwry.NewQQWry("qqwry.dat")
	fmt.Println("version=", qdat.Version())
	fmt.Println(qdat.GetIPLocation(2005104178))
	fmt.Println(qdat.GetIPLocationOfString("119.131.118.50"))
}
```

