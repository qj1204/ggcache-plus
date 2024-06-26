package http

import (
	"fmt"
	"ggcache-plus/distributekv"
	"ggcache-plus/distributekv/consistenthash"
	"ggcache-plus/global"
	"net/http"
	"strings"
	"sync"
)

/*
Because there are other services that may be hosted on a host, it's a good habit to add an extra path,
and most websites have api interfaces that are generally prefixed with api;
*/
const (
	defaultBasePath = "/ggroupcache/"
	defaultReplicas = 50
)

type HTTPPool struct {
	address     string
	basePath    string
	mu          sync.Mutex             // guards peers and httpFetchers
	consHash    *consistenthash.Map    // used to select nodes based on specific keys
	httpClients map[string]*httpClient // keyed by e.g. "http://10.0.0.1:8080"
}

func NewHTTPPool(address string) *HTTPPool {
	return &HTTPPool{
		address:  address,
		basePath: defaultBasePath,
	}
}

// HTTPPool implement HTTP Handler interface
func (p *HTTPPool) Log(format string, v ...interface{}) {
	global.Log.Infof("[Server %s] %s", p.address, fmt.Sprintf(format, v...))
}

func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}

	// print the requested method and requested resource path
	p.Log("%s %s", r.Method, r.URL.Path)

	// prefix/group/key
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request format, expected prefix/group/key", http.StatusBadRequest)
		return
	}
	groupName, key := parts[0], parts[1]

	g := distributekv.GetGroup(groupName)
	if g == nil {
		http.Error(w, "no such group"+groupName, http.StatusBadRequest)
		return
	}

	view, err := g.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	// write value's deep copy
	w.Write(view.ByteSlice())
}

/*
implementing Picker Interface.
function: according to the specific key, select the node and return the HTTP client corresponding to the node.
*/
func (p *HTTPPool) Pick(key string) (distributekv.Fetcher, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	peerAddress := p.consHash.GetTruthNode(key)
	if peerAddress == p.address {
		// upper layer get the value of the key locally after receiving false
		return nil, false
	}

	global.Log.Infof("pick remote peer: %s", peerAddress)
	return p.httpClients[peerAddress], true
}

// rebuilding a consistent hash ring by new peers list
func (p *HTTPPool) UpdatePeers(peers []string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.consHash = consistenthash.New(defaultReplicas, nil)
	p.consHash.Add(peers...)
	p.httpClients = make(map[string]*httpClient, len(peers))

	for _, peer := range peers {
		p.httpClients[peer] = &httpClient{
			baseURL: peer + p.basePath, // such "http://localhost:10001/ggroupcache/"
		}
	}
}

/*
- application/octet-stream 是一种通用的二进制数据类型，用于传输任意类型的二进制数据，没有特定的结构或者格式，可以用于传输图片、音频、视频、压缩文件等任意二进制数据。
- application/json ：用于传输 JSON（Javascript Object Notation）格式的数据，JSON 是一种轻量级的数据交换格式，常用于 Web 应用程序之间的数据传输。
- application/xml：用于传输 XML（eXtensible Markup Language）格式的数据，XML 是一种标记语言，常用于数据的结构化表示和交换。
- text/plain：用于传输纯文本数据，没有特定的格式或者结构，可以用于传输普通文本、日志文件等。
- multipart/form-data：用于在 HTML 表单中上传文件或者二进制数据，允许将表单数据和文件一起传输。
- image/jpeg、image/png、image/gif：用于传输图片数据，分别对应 JPEG、PNG 和 GIF 格式的图片。
- audio/mpeg、audio/wav：用于传输音频数据，分别对应 MPEG 和 WAV 格式的音频
- video/map、video/quicktime：用于传输视频数据，分别对应 MAP4 和 Quicktime 格式的视频。
*/
