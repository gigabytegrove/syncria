package main

import (
 "bytes"
 "encoding/json"
 "fmt"
 "io"
 "log"
 "net"
 "net/http"
 "net/url"
 "strconv"
 "strings"
 "time"
)

func agentLANAddresses(hub string)[]string{
 seen:=map[string]bool{}
 addresses:=[]string{}
 add:=func(ip net.IP){
  if ip==nil{return}
  literal:=ip.String()
  if safeComputerAddress(literal)&&!seen[literal]{
   seen[literal]=true;addresses=append(addresses,literal)
  }
 }
 // Prefer the interface the kernel would use to reach the hub, including
 // hosts with several Ethernet/Wi-Fi/VPN interfaces.
 parsed,_:=url.Parse(hub)
 remoteHost:=parsed.Hostname()
 remotePort:=parsed.Port();if remotePort==""{remotePort="443"}
 conn,err:=net.DialTimeout("udp",net.JoinHostPort(remoteHost,remotePort),3*time.Second)
 if err==nil{
  if addr,ok:=conn.LocalAddr().(*net.UDPAddr);ok{add(addr.IP)}
  conn.Close()
 }
 interfaces,err:=net.InterfaceAddrs()
 if err==nil{
  for _,addr:=range interfaces{
   switch v:=addr.(type){case *net.IPNet:add(v.IP);case *net.IPAddr:add(v.IP)}
  }
 }
 return addresses
}
func announceComputerToHub(hub,code,secret,fingerprint string,port int){
 endpoint:=strings.TrimSuffix(hub,"/")+"/api/computers/enroll"
 client:=&http.Client{Timeout:27*time.Second}
 for attempt:=0;attempt<6;attempt++{
  if attempt>0{time.Sleep(time.Duration(attempt)*2*time.Second)}
  body:=map[string]any{"code":code,"key":secret,"fingerprint":fingerprint,"port":port,"addresses":agentLANAddresses(hub)}
  payload,err:=json.Marshal(body);if err!=nil{return}
  request,err:=http.NewRequest(http.MethodPost,endpoint,bytes.NewReader(payload))
  if err!=nil{return}
  request.Header.Set("Content-Type","application/json")
  response,err:=client.Do(request)
  if err!=nil{
   log.Printf("Syncria enrollment connection attempt %d failed: %v",attempt+1,err);continue
  }
  message,_:=io.ReadAll(io.LimitReader(response.Body,700))
  response.Body.Close()
  if response.StatusCode==http.StatusCreated{
   var result struct{Address string `json:"address"`}
   _=json.Unmarshal(message,&result)
   log.Printf("Syncria agent securely paired at %s",result.Address)
   return
  }
  // An expired or consumed invitation must not be retried indefinitely.
  if response.StatusCode==http.StatusForbidden||response.StatusCode==http.StatusUnauthorized||response.StatusCode==http.StatusConflict{
   log.Printf("Syncria enrollment refused (HTTP %d): %s",response.StatusCode,strings.TrimSpace(string(message)));return
  }
  log.Printf("Syncria enrollment retry %d: HTTP %d %s",attempt+1,response.StatusCode,strings.TrimSpace(string(message)))
 }
 log.Printf("Syncria agent is still available locally but could not auto-pair. Create a fresh enrollment invitation in My Computers.")
}
func agentListenPort(value string)(int,error){
 _,port,err:=net.SplitHostPort(value);if err!=nil{return 0,err}
 n,err:=strconv.Atoi(port);if err!=nil||n<1024||n>65535{return 0,fmt.Errorf("invalid agent port: %s",port)}
 return n,nil
}
