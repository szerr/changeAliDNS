package main

import (
	"errors"
	dns "github.com/alibabacloud-go/alidns-20150109/v2/client"
	openApi "github.com/alibabacloud-go/darabonba-openapi/client"
	"net"
	"os"
	"strings"
)

type configMd struct {
	AccessId   string
	AccessKey  string
	DomainName string
	RR         string
	RegionId   string
	Ip         string
	RecordType string
}

var conf = configMd{
	AccessId:   "<you accessId>",
	AccessKey:  "<you accessKey>",
	DomainName: "<you domain name>",
	RegionId:   "cn-hangzhou",
	RecordType: "A",
}

// 公共请求参数
func globalRequest() ( *dns.Client, error) {
	config := &openApi.Config{}
	// 您的AccessKey ID
	config.AccessKeyId = &conf.AccessId
	// 您的AccessKey Secret
	config.AccessKeySecret = &conf.AccessKey
	// 您的可用区ID
	config.RegionId = &conf.RegionId
	return dns.NewClient(config)
}


func getRecordId() (string, string,  error) {
	client, err := globalRequest()
	if err != nil{
		return "", "", err
	}
	req := &dns.DescribeDomainRecordsRequest{}
	// 主域名
	req.DomainName = &conf.DomainName
	// 主机记录
	req.RRKeyWord = &conf.RR
	resp, err := client.DescribeDomainRecords(req)
	if err != nil{
		return "", "", err
	}
	// 这个主机记录只是个模糊搜索，结果需要迭代一次取出确定的值
	for _, i := range resp.Body.DomainRecords.Record{
		if *i.RR == conf.RR{
			return *i.RecordId, *i.Value, nil
		}
	}
	return "", "", errors.New("This DNS record does not exist:" + conf.RR)
}

func setDNSrecord(recordId , ipAddr string) error {
	client, err := globalRequest()
	if err != nil{
		return err
	}
	// 修改解析记录
	req := &dns.UpdateDomainRecordRequest{}
	// 主机记录
	req.RR = &conf.RR
	// 记录ID
	req.RecordId = &recordId
	// 将主机记录值改为当前主机IP
	req.Value = &ipAddr
	// 解析记录类型
	req.Type = &conf.RecordType
	_, err = client.UpdateDomainRecord(req)
	return err
}

// 获取联网用ip
func GetOutBoundIP()(ip string, err error)  {
	conn, err := net.Dial("udp", "180.101.49.11:80")
	if err != nil {
		return
	}
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	ip = strings.Split(localAddr.String(), ":")[0]
	return
}

func main(){
	var err error
	conf.RR, err = os.Hostname()
	if err != nil{
		panic(err)
	}
	// 阿里云的dns api 需要一个域名id，但是你不知道这个id是啥，所以你需要每次获取一次这个id再设置解析
	recordId, oriIp, err := getRecordId()
	if err != nil{
		panic(err)
	}
	outIp, err := GetOutBoundIP()
	if err != nil{
		panic(err)
	}
	// 如果原ip和目标ip相同，阿里会抛异常
	if outIp != oriIp {
		err = setDNSrecord(recordId, outIp)
		if err != nil {
			panic(err)
		}
	}
}