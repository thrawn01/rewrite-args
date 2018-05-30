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
/usr/bin/ssh worker-n01.prod.us-east-1.postgun.com
```
