/*
   Copyright The Accelerated Container Image Authors

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package server

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/alibaba/accelerated-container-image/pkg/p2p/certificate"
	"github.com/alibaba/accelerated-container-image/pkg/p2p/configure"

	"github.com/elazarl/goproxy"
	log "github.com/sirupsen/logrus"
)

func StartProxyServer(config *configure.DeployConfig, isRun bool) *http.Server {
	proxy := goproxy.NewProxyHttpServer()
	//proxy.Verbose = true
	if config.ProxyConfig.ProxyHTTPS {
		goproxy.GoproxyCa = *certificate.GetRootCA(config.ProxyConfig.CertConfig.CertPath, config.ProxyConfig.CertConfig.KeyPath, config.ProxyConfig.CertConfig.GenerateCert)
		proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	}
	proxy.OnRequest().DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		log.Info("GET URL : ", req.URL.Path)
		match1 := strings.HasPrefix(req.URL.Path, "/v2/")
		match2 := !strings.HasPrefix(req.URL.Path, fmt.Sprintf("/%s/", config.APIKey))
		if match1 && match2 && req.Method == http.MethodGet {
			redirectURL := fmt.Sprintf("%s/%s/%s", config.P2PConfig.MyAddr, config.APIKey, req.URL.String())
			log.Info("Redirect to ", redirectURL)
			header := http.Header{}
			header.Set("Location", redirectURL)
			resp := &http.Response{
				Status:     http.StatusText(http.StatusTemporaryRedirect),
				StatusCode: http.StatusTemporaryRedirect,
				Header:     header,
				Body:       ioutil.NopCloser(bytes.NewBuffer(nil)),
			}
			return req, resp
		}
		return req, nil
	})
	addr := fmt.Sprintf(":%d", config.ProxyConfig.Port)
	server := &http.Server{Addr: addr, Handler: proxy}
	if isRun {
		log.Warnf("Proxy Server Addr ==> %s:%d\n", config.P2PConfig.NodeIP, config.ProxyConfig.Port)
		log.Fatal(http.ListenAndServe(addr, proxy))
	}
	return server
}
