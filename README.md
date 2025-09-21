# Janction Layer One Blockchain - Video Upscaler module


Build docker images

```
docker buildx build \
  --platform linux/amd64,linux/arm64/v8 \
  -t rodrigoa77/upscaler-cpu:latest \
  -t rodrigoa77/upscaler-cpu:amd64 \
  -t rodrigoa77/upscaler-cpu:arm64 \
  --push .
```