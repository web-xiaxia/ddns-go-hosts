# ddns-go-hosts

## wehook url
```
http://host:port/webhook
```

## webhook body
```json
{
    "ipv4Ip":"#{ipv4Addr}",
    "ipv4Hosts":"#{ipv4Domains}",
    "message":{
      "url": "",
      "body": {},
      "headers": {}
    } 
}
```