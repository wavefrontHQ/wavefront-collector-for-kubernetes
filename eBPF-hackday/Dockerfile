FROM quay.io/iovisor/bpftrace:master-vanilla_llvm_clang_glibc2.23 AS bpft

RUN apt-get update
RUN apt-get install -y build-essential bc curl flex bison libelf-dev

ARG UNAME_R=5.10.25-linuxkit
ARG KERNEL_VERSION=5.10.25
ARG MAJOR_VERSION=5

WORKDIR /usr/src/linux
RUN curl -sL "https://www.kernel.org/pub/linux/kernel/v${MAJOR_VERSION}.x/linux-${KERNEL_VERSION}.tar.gz" | tar --strip-components=1 -xzf - -C /usr/src/linux
RUN zcat /proc/config.gz > .config
RUN make ARCH=x86 oldconfig
RUN make ARCH=x86 prepare
RUN mkdir -p /lib/modules/${UNAME_R}
RUN ln -sf /usr/src/linux /lib/modules/${UNAME_R}/source
RUN ln -sf /usr/src/linux /lib/modules/${UNAME_R}/build

WORKDIR /workspace
COPY eBPF-hackday .

# Commands!!!
# ./eBPF-hackday > /dev/null &
# bpftrace -e 'uprobe:/workspace/eBPF-hackday:runtime.mallocgc { printf("%s %s\n", comm, ustack(perf)); }'