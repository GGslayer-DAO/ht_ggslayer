package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func NetLibGetV3(urlpath string, headers interface{}) ([]byte, error) {
	req, err := http.NewRequest("GET", urlpath, nil)
	if err != nil {
		return nil, err
	}
	if headers != nil {
		if maps, ok := headers.(map[string]string); ok {
			for k, v := range maps {
				req.Header.Set(k, v)
			}
		}
	}
	//golang设置代理服务器
	proxy := "http://127.0.0.1:7890"
	proxyAddress, _ := url.Parse(proxy)
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyAddress),
		},
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	result, _ := ioutil.ReadAll(resp.Body)
	return result, nil
}

func NetLibGetV1(url string, headers interface{}) ([]byte, error) {
	client := &http.Client{Timeout: 10 * time.Second} // 超时时间：5秒
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if headers != nil {
		if maps, ok := headers.(map[string]string); ok {
			for k, v := range maps {
				req.Header.Set(k, v)
			}
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	result, _ := ioutil.ReadAll(resp.Body)
	return result, nil
}

// 发送GET请求
// url：请求地址
// response：请求返回的内容
func NetLibGet(url string) ([]byte, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("accept", "application/json")
	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// 发送POST请求
// url：请求地址
// data：POST请求提交的数据
// headers: 请求头内容
// result：返回的内容
func NetLibPost(urlPath string, data interface{}, headers interface{}) ([]byte, error) {
	payload, _ := json.Marshal(data)
	req, _ := http.NewRequest("POST", urlPath, bytes.NewBuffer(payload))
	//golang设置代理服务器
	proxy := "http://127.0.0.1:7890"
	proxyAddress, _ := url.Parse(proxy)
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyAddress),
		},
		Timeout: 10 * time.Second,
	}
	if maps, ok := headers.(map[string]string); ok {
		for k, v := range maps {
			req.Header.Set(k, v)
		}
	}
	res, _ := client.Do(req)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

//NetLibRequest GET,POST公共方法
func NetLibRequest(url, method string, data interface{}, headers interface{}) ([]byte, error) {
	client := &http.Client{Timeout: 5 * time.Second} // 超时时间：5秒
	buffer := &bytes.Buffer{}
	if data != nil {
		jsonStr, _ := json.Marshal(data)
		buffer = bytes.NewBuffer(jsonStr)
	}
	req, err := http.NewRequest(method, url, buffer)
	if err != nil {
		return nil, err
	}
	if maps, ok := headers.(map[string]string); ok {
		for k, v := range maps {
			req.Header.Set(k, v)
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	result, _ := ioutil.ReadAll(resp.Body)
	return result, nil
}

//NetPostForm 发送postform请求
func NetPostForm(urlpath string, datas map[string]string) ([]byte, error) {
	formValues := url.Values{}
	for k, v := range datas {
		formValues.Set(k, fmt.Sprintf("%v", v))
	}
	formDataStr := formValues.Encode()
	//formDataBytes := []byte(formDataStr)
	formBytesReader := strings.NewReader(formDataStr)
	//生成post请求
	client := &http.Client{}
	req, err := http.NewRequest("POST", urlpath, formBytesReader)
	if err != nil {
		return nil, err
	}
	//设置header
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Add("Accept-Encoding", "gzip, deflate, br")
	req.Header.Add("Accept-Language", "zh-CN,zh;q=0.8,en-US;q=0.6,en;q=0.4")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Content-Length", "25")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Cookie", "user_trace_token=20170425200852-dfbddc2c21fd492caac33936c08aef7e; LGUID=20170425200852-f2e56fe3-29af-11e7-b359-5254005c3644; showExpriedIndex=1; showExpriedCompanyHome=1; showExpriedMyPublish=1; hasDeliver=22; index_location_city=%E5%85%A8%E5%9B%BD; JSESSIONID=CEB4F9FAD55FDA93B8B43DC64F6D3DB8; TG-TRACK-CODE=search_code; SEARCH_ID=b642e683bb424e7f8622b0c6a17ffeeb; Hm_lvt_4233e74dff0ae5bd0a3d81c6ccf756e6=1493122129,1493380366; Hm_lpvt_4233e74dff0ae5bd0a3d81c6ccf756e6=1493383810; _ga=GA1.2.1167865619.1493122129; LGSID=20170428195247-32c086bf-2c09-11e7-871f-525400f775ce; LGRID=20170428205011-376bf3ce-2c11-11e7-8724-525400f775ce; _putrc=AFBE3C2EAEBB8730")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36")
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	//Do方法发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	result, _ := ioutil.ReadAll(resp.Body)
	return result, nil
}

func MultipartForm(url string, datas map[string]interface{}) (res []byte, err error) {
	client := &http.Client{Timeout: 5 * time.Second} // 超时时间：5秒
	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)
	req, err := http.NewRequest("POST", url, buf)
	if err != nil {
		return nil, err
	}
	for k, v := range datas {
		w.WriteField(k, fmt.Sprintf("%v", v))
	}
	w.Close()
	req.Header.Set("Content-Type", w.FormDataContentType())
	//Do方法发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	result, _ := ioutil.ReadAll(resp.Body)
	return result, nil
}

func NetLibGetV2(urlPath string, headers interface{}) ([]byte, error) {
	req, err := http.NewRequest("GET", urlPath, nil)
	if err != nil {
		return nil, err
	}
	//golang设置代理服务器
	proxy := "http://127.0.0.1:7890"
	proxyAddress, _ := url.Parse(proxy)
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyAddress),
		},
		Timeout: 10 * time.Second,
	}
	//if headers != nil {
	//	if maps, ok := headers.(map[string]string); ok {
	//		for k, v := range maps {
	//			req.Header.Set(k, v)
	//		}
	//	}
	//}
	req.Header.Set("content-type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	result, _ := ioutil.ReadAll(resp.Body)
	return result, nil
}
