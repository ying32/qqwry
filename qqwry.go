/********************************************************

       纯真IP数据库查询
       by:ying32
       2016/5/13
       QQ: 1444386932
--------------------------------------------------------

       未测试多种带有重定义或未重定义的IP
       如果查询出错，可发邮件给我，告知改进
       或者由您改进，但记得发我一份

       纯真网络：
         http://www.cz88.net/
       纯真IP数据格式：
         http://lumaqq.linuxsir.org/
                     article/qqwry_format_detail.html

********************************************************/

package qqwry

import (
	"bytes"
	"code.google.com/p/mahonia"
	"encoding/binary"
	"fmt"
	"net"
	"os"
)

// 重定义模式
const (
	REDIRECT_MODE_1 = 1
	REDIRECT_MODE_2 = 2
)

var (
	// 全局数据，只一次
	dataReader *bytes.Reader
	// 标识"qqwry.dat"文件是否初始
	isInitReader bool
)

type QQWry struct {
	RecordCount uint32
	FirstRecord uint32
}

// 模拟构造函数
func NewQQWry(fileName string) *QQWry {
	// 只初始一次文件
	if !isInitReader {
		initFile(fileName)
		isInitReader = true
	}
	var rf, rl, rc uint32
	binary.Read(dataReader, binary.LittleEndian, &rf)
	binary.Read(dataReader, binary.LittleEndian, &rl)
	rc = (rl - rf) / 7
	return &QQWry{rc, rf}
}

// 获取版本信息
func (self *QQWry) Version() string {
	if self.RecordCount > 0 {
		offset, _, _ := self.getIPOffset(self.RecordCount, 1)
		if offset > 0 {
			return self.getIpLocationOfOffset(offset)
		}
	}
	return "未知区域"
}

// 读取一个3字节并转为整数
func (self *QQWry) read3byteuint64() int64 {
	temp := make([]byte, 3)
	dataReader.Read(temp)
	return int64(uint32(temp[0]) | uint32(temp[1])<<8 | uint32(temp[2])<<16)
}

// 获取当前Reader位置
func (self *QQWry) dataReaderPosition() int64 {
	return dataReader.Size() - int64(dataReader.Len())
}

// AnsiToUtf8的转换
func (self *QQWry) ansiToUtf8(bytes []byte) string {
	decoder := mahonia.NewDecoder("gbk")
	if res, ok := decoder.ConvertStringOK(string(bytes)); ok {
		return res
	}
	return ""
}

// 字符串ip转整型
func (self *QQWry) StrToIP(ipstr string) uint32 {
	ip := net.ParseIP(ipstr)
	return binary.BigEndian.Uint32(ip.To4())
}

// 整型转ip字符，ipv4
func (self *QQWry) IPToStr(ip uint32) string {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, ip)
	return fmt.Sprintf("%d.%d.%d.%d", b[0], b[1], b[2], b[3])
}

// 获取ip偏移位置
func (self *QQWry) getIPOffset(recNum uint32, flag int) (int64, uint32, uint32) {
	var begin, end uint32
	var result int64 = int64(self.FirstRecord + recNum*7)
	dataReader.Seek(result, 0)
	binary.Read(dataReader, binary.LittleEndian, &begin)
	if flag == 1 {
		result = self.read3byteuint64()
		dataReader.Seek(result, 0)
		binary.Read(dataReader, binary.LittleEndian, &end)
	}
	return result, begin, end
}

// 读取区域信息
func (self *QQWry) readAreaStrOfOffset(offset int64) string {
	dataReader.Seek(offset, 0)
	b, _ := dataReader.ReadByte()
	if b == REDIRECT_MODE_1 || b == REDIRECT_MODE_2 {
		dataReader.Seek(offset+1, 0)
		AreaOffset := self.read3byteuint64()
		if AreaOffset != 0 {
			return self.readStrOfOffset(AreaOffset)
		}
	} else {
		return self.readStrOfOffset(offset)
	}
	return "未知区域"
}

// 读指定位置的字符串
func (self *QQWry) readStrOfOffset(offset int64) string {
	dataReader.Seek(offset, 0)
	strbytes := make([]byte, 0) // 字节数组
	b, _ := dataReader.ReadByte()
	for b != 0 {
		strbytes = append(strbytes, b)
		b, _ = dataReader.ReadByte()
	}
	return self.ansiToUtf8(strbytes)
}

// 读取指定位置ip的归属字符串
func (self *QQWry) getIpLocationOfOffset(offset int64) string {
	dataReader.Seek(offset+4, 0)
	var CountryName, AreaName string
	b, _ := dataReader.ReadByte()

	switch b {
	case REDIRECT_MODE_1:
		Countryoffset := self.read3byteuint64()
		dataReader.Seek(Countryoffset, 0)
		b, _ := dataReader.ReadByte()
		if b == REDIRECT_MODE_2 {
			CountryName = self.readStrOfOffset(self.read3byteuint64())
			dataReader.Seek(Countryoffset+4, 0)
		} else {
			CountryName = self.readStrOfOffset(Countryoffset)
		}
		AreaName = self.readAreaStrOfOffset(self.dataReaderPosition())

	case REDIRECT_MODE_2:
		CountryName = self.readStrOfOffset(self.read3byteuint64())
		AreaName = self.readAreaStrOfOffset(offset + 8)

	default:
		CountryName = self.readStrOfOffset(self.dataReaderPosition() - 1)
		AreaName = self.readAreaStrOfOffset(self.dataReaderPosition())

	}
	return CountryName + AreaName
}

// 读取指定ip字符串位置
func (self *QQWry) GetIPLocation(ip uint32) string {
	if self.RecordCount == 0 || ip <= 0 {
		return ""
	}
	var min, max, mid, beginIP, endIP uint32
	max = self.RecordCount - 1
	for min <= max {
		mid = (min + max) / 2
		_, beginIP, endIP = self.getIPOffset(mid, 0)
		if ip == beginIP {
			max = mid
			break
		} else if ip > beginIP {
			min = mid + 1
		} else {
			max = mid - 1
		}
	}
	offset, beginIP, endIP := self.getIPOffset(max, 1)
	if beginIP <= ip && endIP >= ip {
		return self.getIpLocationOfOffset(offset)
	}
	return "***友情提示，未知IP***"
}

// 查找字符串ip
func (self *QQWry) GetIPLocationOfString(ipstr string) string {
	return self.GetIPLocation(self.StrToIP(ipstr))
}

// 初始文件, 只初始化一次
func initFile(fileName string) {
	if isInitReader {
		return
	}
	// init data
	dataStream, err := os.Open(fileName)
	if err != nil {
		panic("打开文件错误")
	}
	defer dataStream.Close()
	fileInfo, err := dataStream.Stat()
	if err != nil {
		panic("获取文件信息错误")
	}
	tempbuffer := make([]byte, fileInfo.Size())
	_, err = dataStream.Read(tempbuffer)
	if err != nil {
		panic("读取错误")
	}
	dataReader = bytes.NewReader(tempbuffer)
}
