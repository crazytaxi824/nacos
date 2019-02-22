package nacos

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/dxvgef/go-lib/httplib"
)

// 查找服务地址
type respResult struct {
	Hosts []hostsSlice `json:"hosts"`
}

type hostsSlice struct {
	IP   string `json:"ip"`
	Port int    `json:"port"`
}

// SearchService 通过服务名字查找服务，返回 服务[ip:port], errMsg, err
func (ss *SaasService) SearchService(serviceName string, reqURL string) (paths []string, err error) {
	// 发现服务
	// "http://192.168.2.250:8848/nacos/v1/ns/instance/list"
	req := httplib.Get(reqURL)
	req.Param("serviceName", serviceName)

	resp, err := req.Bytes()
	if err != nil {
		return nil, err
	}

	var hosts respResult
	err = json.Unmarshal(resp, &hosts)
	if err != nil {
		return nil, errors.New("json Unmarshal error: " + string(resp))
	}

	for _, v := range hosts.Hosts {
		paths = append(paths, v.IP+":"+strconv.Itoa(v.Port))
	}

	return paths, nil
}
