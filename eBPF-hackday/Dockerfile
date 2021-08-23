FROM docker/for-desktop-kernel:5.10.25-6594e668feec68f102a58011bb42bd5dc07a7a9b AS ksrc

FROM ubuntu:19.10

WORKDIR /
COPY --from=ksrc /kernel-dev.tar /
RUN tar xf kernel-dev.tar && rm kernel-dev.tar

RUN  apt-get update \
  && apt-get install -y wget \
  && apt-get install -y bpftrace \
  && apt-get install -y kmod python3-bpfcc

WORKDIR /workspace
COPY eBPF-hackday .
RUN wget https://github.com/iovisor/bpftrace/releases/download/v0.13.0/bpftrace
RUN chmod +x bpftrace

CMD mount -t debugfs debugfs /sys/kernel/debug && /bin/bash

#RUN apt-get install -y libbpfcc-dev

#RUN apt-get update
#RUN apt-get install -y \
#  bison \
#  cmake \
#  flex \
#  g++ \
#  git \
#  libelf-dev \
#  zlib1g-dev \
#  libfl-dev \
#  systemtap-sdt-dev \
#  binutils-dev \
#  libcereal-dev \
#  llvm-7-dev \
#  llvm-7-runtime \
#  libclang-7-dev \
#  clang-7 \
#  libgtest-dev \
##  libgmock-dev \
#  asciidoctor
#
#RUN git clone https://github.com/iovisor/bpftrace
#RUN mkdir bpftrace/build; cd bpftrace/build;
#RUN cmake -DCMAKE_BUILD_TYPE=Release ..
#RUN make -j8
#RUN make install