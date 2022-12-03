package main


import (
	"fmt"
	"net/http/httputil"
	"net/http"
	"net/url"
	"os"

)

type Server interface{
	Address() string
	IsAlive() bool
	Serve(r *http.Request, w http.ResponseWriter)
}


type LoadBalancer struct{
	Port  string
	Servers  []Server
	RoundRobbinCount  int
}

func newLoadBalancer(port string, servers []Server) *LoadBalancer{

	return &LoadBalancer{
		Port: port,
		Servers: servers,
		RoundRobbinCount: 0,
	}
}


type SimpleServer struct{
	Addr string
	Proxy *httputil.ReverseProxy
}

func simpleServer(addr string) *SimpleServer {
	parseUrl, err := url.Parse(addr)
	handleErr(err)

	return &SimpleServer{
		Addr: addr,
		Proxy: httputil.NewSingleHostReverseProxy(parseUrl),
	}

}

func handleErr(err error){
	 if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	 }
}

func (s *SimpleServer) Address() string{
	 return s.Addr
	}

func (s *SimpleServer) IsAlive() bool{return true}

func (s *SimpleServer) Serve(req *http.Request, rw http.ResponseWriter){
	s.Proxy.ServeHTTP(rw, req)
}


func (l *LoadBalancer) getNextAvailableServer() Server {
	server := l.Servers[l.RoundRobbinCount % len(l.Servers)]
	if !server.IsAlive(){
		l.RoundRobbinCount++
		server = l.Servers[l.RoundRobbinCount % len(l.Servers)]
	}
	l.RoundRobbinCount++
	return server
}

func (l *LoadBalancer) ServeProxy( w http.ResponseWriter, r *http.Request){
	targetServer := l.getNextAvailableServer()
	fmt.Println("Forward to the address: ", targetServer.Address())
	targetServer.Serve(r, w)
}


func main(){
	servers := []Server{
		simpleServer("https://facebook.com/"),
		simpleServer("https://www.google.com"),
		simpleServer("https://duckduckgo.com"),
	}

	lb := newLoadBalancer("8080", servers)

	handleRedirect := func( w http.ResponseWriter, r *http.Request){
		lb.ServeProxy(w, r)
	}

	http.HandleFunc("/", handleRedirect)
	
	fmt.Println("server started at localhost:", lb.Port)
	http.ListenAndServe(":" +lb.Port, nil)

}