### 依赖

golang > 1.17

### 使用容器环境

```
docker run -it -v $(pwd):/baidingyun /root/.kube:/root/.kube golang:1.18.6 bash
```

### 设置 go proxy
```
go env -w GO111MODULE=on
go env -w  GOPROXY=https://goproxy.cn,direct
```