#### If the following is true
```
ln -s $GOPATH/bin/rewrite-args $HOME/bin/ssh
export PATH="$HOME/bin:$PATH"
```

#### and ~/.rewrite-args.conf contains
```json
{
  "items": [
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

You can setup additional links for scp or any other commands
and any aliases you have will work with rewrite-args
