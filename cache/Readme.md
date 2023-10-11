# Simple proxy cache


## The forward proxy cache for static web content:

```
               network            proxy             cache
             .---------.       .---------.       .---------.
respond <----|- - < - -|---<---|- - < - -|---<---|- < -.   |
request ---->|- - > - -|--->---|- -,- > -|--->---|- > -|   |
             |         |       |   |(*)  |       |     |   |
             |    ,-< -|---<---|< -'     |       |     |   |
             |    , ,->|--->---|- - > - -|--->---|- > -'   |
             `----+-+--´       `---------´       `---------´
                  ' '
                  '_'
                website

(*) When the data is not in the cache, the website will be requested and is directly stored in the cache.
(*) Where "network" may be anything (LAN/WAN/...).
```

Run this command from repo root to start the foward proxy cache:
```
make fordward
```
Try these links:
- http://localhost:8080
- http://localhost:8080/features/copilot
- http://localhost:8080/golang/go
- http://localhost:8080/nginx/nginx

Static data will be cache into `cache` folder.