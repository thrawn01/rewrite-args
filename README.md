## Synopsis

#### If the following is true
```
alias ssh='rewrite-args ssh -X'
```

#### and ~/.rewrite-args.conf contains
```json
{
  "debug": false,
  "rewrites": [
    {
      "match": ".use1",
      "replace": ".prod.us-east-1.postgun.com"
    }
  ]
}
```

#### Given the following command
```
ssh worker-n01.use1
```

#### Will expand too
```
/usr/bin/ssh -X worker-n01.prod.us-east-1.postgun.com
```

## Installation

Download the latest binary [release](https://github.com/thrawn01/rewrite-args/releases)

**OR**

`go install github.com/thrawn01/rewrite-args`
