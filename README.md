DomInscpect: A small script to make Dom based vulnerability detection easier based on matchers

### Usage:

```bash
cat urls.txt | ./dominspect
```

### Command line flags:

```
Usage of ./dominspect:
  -config string
        path to the JSON configuration file (default "~/dominspect.json")
  -p int
        number of concurrent executions (default 5)
```

* Default location of matcher file : `~/dominspect.json`

> Example matchers

```json
[
    {
        "key": "XSS",
        "value": "<script>alert()</script>"
    }
]
```
