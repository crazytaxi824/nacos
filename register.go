package nacos

import (
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"time"

	"github.com/dxvgef/go-lib/httplib"
)

// SaasService SaasService
type SaasService struct {
	err error
}

// SaasOption SaasOption
type SaasOption struct {
	// 必填项
	CenterURL   string // 注册地址,最后不要/
	IP          string // 自己的ip地址
	Port        string // 自己的端口地址
	ServiceName string // 自己的服务名称
	HeartBeats  int    // 心跳包间隔时间，单位s

	// 以下为选填,我也不知道干嘛的
	NamespaceID string
	Weight      int
	Enable      bool
	Healthy     bool
	Metadata    map[string]interface{} //元数据
	ClusterName string                 // 集群名称
	Scheduled   bool
}

// ResigterService 注册服务
func (ss *SaasService) ResigterService(option *SaasOption) error {
	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
		}
	}()

	// saasOption = *option
	// 注册事例
	req := httplib.Put(option.CenterURL)
	req.Param("ip", option.IP)
	req.Param("port", option.Port)
	req.Param("serviceName", option.ServiceName)

	// 可选
	if option.Weight != 0 {
		req.Param("weight", strconv.Itoa(option.Weight))
	}

	if option.Enable {
		req.Param("enable", "true")
	}

	if option.Healthy {
		req.Param("healthy", "true")
	}

	b, err := json.Marshal(option.Metadata)
	if err != nil {
		return err
	}

	req.Param("metadata", string(b))

	if option.ClusterName != "" {
		req.Param("clusterName", option.ClusterName)
	}

	if option.NamespaceID != "" {
		req.Param("namespaceId", option.NamespaceID)
	}

	resp, err := req.String()
	if err != nil {
		return err
	}

	// log.Println(resp)

	if resp != "ok" && resp != "OK" {
		return errors.New("服务注册错误：" + resp)
	}

	// 心跳包
	ticker := time.NewTicker(time.Duration(option.HeartBeats) * time.Second)
	go func(ticker *time.Ticker, option *SaasOption) {
		ss.HeartBeats(ticker, option)
	}(ticker, option)

	return ss.err
}

// HeartBeats 心跳服务，需要传倒计时 ticker
// 需要在携程中开启
func (ss *SaasService) HeartBeats(ticker *time.Ticker, option *SaasOption) {
	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
		}
	}()

	// 心跳包
	for {
		select {
		case <-ticker.C:
			beats := httplib.Put(option.CenterURL + "/beat")
			beats.Param("serviceName", option.ServiceName)

			beatsMap := make(map[string]interface{})
			beatsMap["metadata"] = option.Metadata
			beatsMap["serviceName"] = option.ServiceName
			beatsMap["ip"] = option.IP
			beatsMap["port"] = option.Port

			if option.Weight != 0 {
				beatsMap["weight"] = option.Weight
			}

			if option.ClusterName != "" {
				beatsMap["cluster"] = option.ClusterName
			}

			if option.Scheduled {
				beatsMap["scheduled"] = option.Scheduled
			}

			b, err := json.Marshal(beatsMap)
			if err != nil {
				ss.err = err
				return
			}

			beats.Param("beat", string(b))

			resp, err := beats.String()
			if err != nil {
				ss.err = err
				return
			}

			if resp != "{\"clientBeatInterval\":5000}" && resp != "ok" && resp != "OK" {
				ss.err = errors.New("心跳包错误：" + resp)
				return
			}
		}
	}
}

// ResigterService2 注册服务
func resigterService2() error {
	// 注册事例
	req := httplib.Put("http://192.168.2.250:8848/nacos/v1/ns/instance")
	req.Param("ip", "192.168.2.103")
	req.Param("port", "10082")
	// req.Param("namespaceId", "public")
	// req.Param("weight", "1.0")
	// req.Param("enable", "true")
	// req.Param("healthy", "true")
	// req.Param("metadata", "{}")
	// req.Param("clusterName", "DEFAULT")
	req.Param("serviceName", "test_sms")

	resp, err := req.String()
	if err != nil {
		return err
	}

	// log.Println(resp)

	if resp != "ok" && resp != "OK" {
		return errors.New("服务注册错误：" + resp)
	}

	return nil
}
