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
