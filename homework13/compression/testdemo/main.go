package main

import (
	"bytes"
	"compress/gzip"
)

///test demo--

func main() {
	//text2 := "你好啊assa 啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊sx啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊啊"
	//srcbs := []byte(text)
	//fmt.Println(srcbs)
	//
	//buffer := bytes.NewBuffer(nil)
	//gzw := gzip.NewWriter(buffer)
	//gzw.Write(srcbs)
	//gzw.Close()
	//fmt.Println(gzw)

	//fmt.Println([]byte(text2), len([]byte(text2)))
	//res, ok := Encode([]byte(text2))
	//fmt.Println(res, ok, len(res))
	//r, ok := Decode(res)
	//fmt.Println(string(r))
	//bs := make([]byte, 5)
	//copy(bs, []byte("nihaoaaa"))
	//fmt.Println(string(bs))
}

func Encode(input []byte) ([]byte, error) {
	// 创建一个新的 byte 输出流
	var buf bytes.Buffer
	// 创建一个新的 gzip 输出流
	gzipWriter := gzip.NewWriter(&buf)
	// 将 input byte 数组写入到此输出流中
	_, err := gzipWriter.Write(input)
	if err != nil {
		_ = gzipWriter.Close()
		return nil, err
	}
	if err := gzipWriter.Close(); err != nil {
		return nil, err
	}
	// 返回压缩后的 bytes 数组
	return buf.Bytes(), nil
}

func Decode(input []byte) ([]byte, error) {
	// 创建一个新的 gzip.Reader
	bytesReader := bytes.NewReader(input)
	gzipReader, err := gzip.NewReader(bytesReader)
	if err != nil {
		return nil, err
	}
	defer func() {
		// defer 中关闭 gzipReader
		_ = gzipReader.Close()
	}()
	buf := new(bytes.Buffer)
	// 从 Reader 中读取出数据
	if _, err := buf.ReadFrom(gzipReader); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
