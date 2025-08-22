# Dockerfile
FROM ubuntu:22.04
ENV DEBIAN_FRONTEND=noninteractive

# Base deps + OpenCL headers (to satisfy CMake); we will force CPU at runtime.
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential cmake git ca-certificates \
    ffmpeg \
    libopencv-dev \
    ocl-icd-opencl-dev opencl-headers \
    && rm -rf /var/lib/apt/lists/*

# Clone & build waifu2x-converter-cpp
ARG W2X_REF=v5.3.4
RUN git clone --depth 1 --branch ${W2X_REF} https://github.com/DeadSix27/waifu2x-converter-cpp.git /opt/w2x && \
    cd /opt/w2x && \
    cmake -DCMAKE_BUILD_TYPE=Release \
    -DENABLE_GUI=OFF \
    . && \
    make -j"$(nproc)" && \
    strip waifu2x-converter-cpp && \
    ln -s /opt/w2x/waifu2x-converter-cpp /usr/local/bin/waifu2x-converter-cpp

# Install models from the repo into a predictable path
RUN mkdir -p /usr/local/share/waifu2x-converter-cpp && \
    cp -r /opt/w2x/models_rgb /usr/local/share/waifu2x-converter-cpp/models_rgb

# Entrypoint script
COPY upscale_cpu.sh /usr/local/bin/upscale_cpu.sh
RUN chmod +x /usr/local/bin/upscale_cpu.sh

WORKDIR /work
ENTRYPOINT ["/usr/local/bin/upscale_cpu.sh"]
