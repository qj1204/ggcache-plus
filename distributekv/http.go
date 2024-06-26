package distributekv

//
//import (
//	"fmt"
//	"ggcache-plus/distributekv/consistenthash"
//	pb "ggcache-plus/distributekv/ggmemcachedpb"
//	"google.golang.org/protobuf/proto"
//	"io/ioutil"
//	"log"
//	"net/http"
//	"net/url"
//	"strings"
//	"sync"
//)
//
//const (
//	defaultBasePath = "/_distributekv/"
//	defaultReplicas = 50
//)
//
//// HTTPPool implements PeerPicker for a pool of HTTP peers.
//type HTTPPool struct {
//	selfAddr     string                  // 记录自己的地址, e.g. "http://localhost:8001"
//	basePath     string                  // 节点间通讯地址的前缀
//	mu           sync.Mutex              // guards peers and httpFetchers
//	peers        *consistenthash.Map     // 一致性哈希算法的实例
//	httpFetchers map[string]*httpFetcher // 每一个远程节点地址对应一个 httpFetcher
//}
//
//// NewHTTPPool initializes an HTTP pool of peers.
//func NewHTTPPool(selfAddr string) *HTTPPool {
//	return &HTTPPool{
//		selfAddr: selfAddr,
//		basePath: defaultBasePath,
//	}
//}
//
//// Log info with server name
//func (p *HTTPPool) Log(format string, v ...interface{}) {
//	log.Printf("[Server %s] %s", p.selfAddr, fmt.Sprintf(format, v...))
//}
//
//// ServeHTTP handle all http requests
//func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
//	if !strings.HasPrefix(r.URL.Path, p.basePath) {
//		panic("HTTPPool serving unexpected path: " + r.URL.Path)
//	}
//	p.Log("%s %s", r.Method, r.URL.Path)
//	// /<basepath>/<groupname>/<key> required
//	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
//	if len(parts) != 2 {
//		http.Error(w, "bad request", http.StatusBadRequest)
//		return
//	}
//
//	groupName := parts[0]
//	group := GetGroup(groupName)
//	if group == nil {
//		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
//		return
//	}
//	key := parts[1]
//	view, err := group.Get(key)
//	if err != nil {
//		http.Error(w, err.Error(), http.StatusInternalServerError)
//		return
//	}
//
//	// Write the value to the response body as a proto message.
//	body, err := proto.Marshal(&pb.Response{Value: view.ByteSlice()})
//	if err != nil {
//		http.Error(w, err.Error(), http.StatusInternalServerError)
//		return
//	}
//
//	w.Header().Set("Content-Type", "application/octet-stream")
//	w.Write(body)
//}
//
//// httpFetcher HTTP客户端结构体
//type httpFetcher struct {
//	baseURL string
//}
//
//// Fetch 通过http客户端获取远程节点数据
//func (h *httpFetcher) Fetch(in *pb.Request, out *pb.Response) error {
//	u := fmt.Sprintf(
//		"%v%v/%v",
//		h.baseURL, // e.g. "http://localhost:8001/_ggmemcached/"
//		url.QueryEscape(in.GetGroup()),
//		url.QueryEscape(in.GetKey()),
//	)
//	res, err := http.Get(u)
//	if err != nil {
//		return err
//	}
//	defer res.Body.Close()
//
//	if res.StatusCode != http.StatusOK {
//		return fmt.Errorf("server returned: %v", res.Status)
//	}
//
//	bytes, err := ioutil.ReadAll(res.Body)
//	if err != nil {
//		return fmt.Errorf("reading response body: %v", err)
//	}
//	if err = proto.Unmarshal(bytes, out); err != nil {
//		return fmt.Errorf("decoding response body: %v", err)
//	}
//
//	return nil
//}
//
//// Set 实例化一致性哈希算法，并添加传入的节点
//func (p *HTTPPool) Set(peers ...string) {
//	p.mu.Lock()
//	defer p.mu.Unlock()
//	p.peers = consistenthash.New(defaultReplicas, nil)
//	p.peers.Add(peers...)
//	p.httpFetchers = make(map[string]*httpFetcher, len(peers))
//	for _, peer := range peers {
//		p.httpFetchers[peer] = &httpFetcher{baseURL: peer + p.basePath}
//	}
//}
//
//// Pick 根据具体的 key 选择节点，返回节点对应的 HTTP 客户端
//func (p *HTTPPool) Pick(key string) (Fetcher, bool) {
//	p.mu.Lock()
//	defer p.mu.Unlock()
//	if peer := p.peers.Get(key); peer != "" && peer != p.selfAddr {
//		p.Log("Pick peer %s", peer)
//		return p.httpFetchers[peer], true
//	}
//	return nil, false
//}
