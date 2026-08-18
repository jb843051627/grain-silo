# benzhi Docker 构建说明

本项目使用 `benzhi.Dockerfile` 构建标准 Docker 镜像。

## 构建

```bash
# 构建指定架构
./build_benzhi_docker.sh grain-silo-bug-1 linux/amd64
./build_benzhi_docker.sh grain-silo-bug-1 linux/arm64

# 或直接用 docker buildx
docker buildx build --platform linux/amd64 -t benzhi/grain-silo-bug-1:latest -f benzhi.Dockerfile .
```

## 运行

```bash
docker run -p 8080:8080 benzhi/grain-silo-bug-1:latest
```

## 说明

- 基础镜像：`golang:1.22-bookworm`
- 代理：`GOPROXY=https://goproxy.cn,direct`
- 工具链：`GOTOOLCHAIN=local`
- 数据库：SQLite 文件模式，无外部依赖
