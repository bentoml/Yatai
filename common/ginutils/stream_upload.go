package ginutils

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"strconv"

	"github.com/pkg/errors"
)

// 解析多个文件上传中，每个具体的文件的信息
type FileHeader struct {
	ContentDisposition string
	Name               string
	FileName           string // 文件名
	ContentType        string
	ContentLength      int64
}

const (
	ContentDisposition = "Content-Disposition: "
	NAME               = "name=\""
	FILENAME           = "filename=\""
	ContentType        = "Content-Type: "
	ContentLength      = "Content-Length: "
)

var (
	boundaryHeaderSeparator  = []byte("\r\n")
	headerContentSeparator   = []byte("\r\n\r\n")
	contentBoundarySeparator = []byte("\r\n")
)

// 解析描述文件信息的头部
// @return FileHeader 文件名等信息的结构体
// @return bool 解析成功还是失败
func ParseFileHeader(h []byte) (FileHeader, bool) {
	arr := bytes.Split(h, boundaryHeaderSeparator)
	var outHeader FileHeader
	outHeader.ContentLength = -1
	for _, item := range arr {
		// nolint: gocritic
		if bytes.HasPrefix(item, []byte(ContentDisposition)) {
			l := len(ContentDisposition)
			arr1 := bytes.Split(item[l:], []byte("; "))
			outHeader.ContentDisposition = string(arr1[0])
			if bytes.HasPrefix(arr1[1], []byte(NAME)) {
				outHeader.Name = string(arr1[1][len(NAME) : len(arr1[1])-1])
			}
			l = len(arr1[2])
			if bytes.HasPrefix(arr1[2], []byte(FILENAME)) && arr1[2][l-1] == 0x22 {
				outHeader.FileName = string(arr1[2][len(FILENAME) : l-1])
			}
		} else if bytes.HasPrefix(item, []byte(ContentType)) {
			l := len(ContentType)
			outHeader.ContentType = string(item[l:])
		} else if bytes.HasPrefix(item, []byte(ContentLength)) {
			l := len(ContentLength)
			s := string(item[l:])
			contentLength, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				log.Printf("content length error:%s", string(item))
				return outHeader, false
			} else {
				outHeader.ContentLength = contentLength
			}
		} else {
			log.Printf("unknown:%s\n", string(item))
		}
	}
	if len(outHeader.FileName) == 0 {
		return outHeader, false
	}
	return outHeader, true
}

// 从流中一直读到文件的末位
// @return []byte 没有写到文件且又属于下一个文件的数据
// @return int 没有写到文件且又属于下一个文件的数据的长度
// @return bool 是否已经读到流的末位了
// @return error 是否发生错误
func GetFileContentFromUploadStream(readData []byte, readTotal int, boundary []byte, stream io.ReadCloser, target io.WriteCloser) ([]byte, int, bool, error) {
	buf := make([]byte, 1024*4)

	// nolint: gocritic
	fileContentEndBoundary := append(contentBoundarySeparator, append(append([]byte("--"), boundary...), []byte("--")...)...)
	fileContentEndBoundaryLen := len(fileContentEndBoundary)

	reachEnd := false
	for !reachEnd {
		readLen, err := stream.Read(buf)
		if err != nil {
			if !errors.Is(err, io.EOF) && readLen <= 0 {
				return nil, 0, true, err
			}
			reachEnd = true
		}
		// todo: 下面这一句很蠢，值得优化
		copy(readData[readTotal:], buf[:readLen]) // 追加到另一块 buffer，仅仅只是为了搜索方便
		readTotal += readLen
		if readTotal < fileContentEndBoundaryLen {
			continue
		}
		fileContentEndIndex := bytes.Index(readData[:readTotal], fileContentEndBoundary)
		if fileContentEndIndex >= 0 {
			_, _ = target.Write(readData[:fileContentEndIndex])
			return readData[fileContentEndIndex:], readTotal - fileContentEndIndex, reachEnd, nil
		}

		_, _ = target.Write(readData[:readTotal-fileContentEndBoundaryLen])
		copy(readData[0:], readData[readTotal-fileContentEndBoundaryLen:])
		readTotal = fileContentEndBoundaryLen
	}
	_, _ = target.Write(readData[:readTotal])
	return nil, 0, reachEnd, nil
}

// 解析表单的头部
// @param read_data 已经从流中读到的数据
// @param read_total 已经从流中读到的数据长度
// @param boundary 表单的分割字符串
// @param stream 输入流
// @return FileHeader 文件名等信息头
//
//				[]byte 已经从流中读到的部分
//	        int 已经从流中读取到的大小
//				error 是否发生错误
func GetFileHeaderFromUploadStream(readData []byte, readTotal int, boundary []byte, stream io.ReadCloser) (FileHeader, []byte, int, error) {
	buf := make([]byte, 1024*4)
	foundBoundary := false
	boundaryIndex := -1
	boundaryLen := len(boundary)
	headerContentSeparatorLen := len(headerContentSeparator)

	var fileHeader FileHeader
	for {
		readLen, err := stream.Read(buf)
		if err != nil && !errors.Is(err, io.EOF) {
			return fileHeader, nil, 0, err
		}
		if readTotal+readLen > cap(readData) {
			return fileHeader, nil, 0, fmt.Errorf("not found boundary")
		}
		copy(readData[readTotal:], buf[:readLen])
		readTotal += readLen
		if !foundBoundary {
			boundaryIndex = bytes.Index(readData[:readTotal], boundary)
			if -1 == boundaryIndex {
				continue
			}
			foundBoundary = true
		}
		fileHeaderStartIndex := boundaryIndex + boundaryLen
		fileHeaderEndIndex := bytes.Index(readData[fileHeaderStartIndex:readTotal], headerContentSeparator)
		if fileHeaderEndIndex == -1 {
			continue
		}
		fileHeaderEndIndex += fileHeaderStartIndex
		var ret bool
		fileHeader, ret = ParseFileHeader(readData[fileHeaderStartIndex:fileHeaderEndIndex])
		if !ret {
			return fileHeader, nil, 0, fmt.Errorf("ParseFileHeader fail: %s", string(readData[fileHeaderStartIndex:fileHeaderEndIndex]))
		}
		fileContentStartIndex := fileHeaderEndIndex + headerContentSeparatorLen
		return fileHeader, readData[fileContentStartIndex:], readTotal - fileContentStartIndex, nil
	}
}
